package surfboard_service_order

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	surfboard_service "github.com/DowLucas/gin-ticket-release/pkg/services/surfboard"
	"gorm.io/gorm"
)

type SurfboardCreateOrderService struct {
	db           *gorm.DB
	client       surfboard_service.SurfboardClient
	orderService *OrderService
}

func NewSurfboardCreateOrderService(db *gorm.DB) *SurfboardCreateOrderService {
	return &SurfboardCreateOrderService{db: db,
		client:       surfboard_service.NewSurfboardClient(),
		orderService: NewOrderService()}
}

func (sos *SurfboardCreateOrderService) CreateOrder(ticketIDs []uint,
	network *models.Network,
	event *models.Event,
	user *models.User) (*models.Order, error) {
	// Recieves a list of ticket IDs and creates an order for them
	// The order will be created in the database and should return a link to the user to pay for the order
	tx := sos.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	tickets, err := models.GetTicketsByIDs(tx, ticketIDs)
	if err != nil {
		return nil, err
	}

	for _, ticket := range tickets {
		if ticket.TicketOrder.TicketRelease.EventID != int(event.ID) {
			tx.Rollback()
			return nil, errors.New("ticket does not belong to event")
		}
	}

	// After we have the tickets we send them to the surfboard order service CreateNewOrder method
	order, err := sos.createOrder(tx, &event.Terminal, network, user, tickets)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, ticket := range tickets {
		// Update order_id on ticket
		if err := tx.Model(&ticket).Update("order_id", order.ID).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	return order, nil
}

type OrderResponse struct {
	Status string `json:"status"`
	Data   struct {
		OrderID         string `json:"orderId"`
		PaymentPageLink string `json:"paymentPageLink"`
	} `json:"data"`
	Message string `json:"message"`
}

type Data struct {
}

func (sos *SurfboardCreateOrderService) createOrder(
	tx *gorm.DB,
	terminal *models.StoreTerminal,
	network *models.Network,
	user *models.User,
	tickets []models.Ticket) (*models.Order, error) {

	// Checks to see so that everything is preloaded
	if terminal.TerminalID == "" {
		return nil, errors.New("terminal not found")
	}

	if network.Merchant.MerchantID == "" {
		return nil, errors.New("merchant not found")
	}

	merchant := network.Merchant

	orderLines := sos.generateOrderLines(tickets)

	orderData := OrderRequest{
		TerminalID: terminal.TerminalID,
		Type:       "purchase",
		OrderLines: orderLines,
		ControlFunctions: ControlFunctions{
			CancelPreviousPendingOrder: true,
		},
	}

	orderBytes, err := json.Marshal(orderData)
	if err != nil {
		panic(err)
	}

	response, err := sos.orderService.createOrder(merchant.MerchantID, orderBytes)
	if err != nil {
		return nil, err
	}

	body, _ := io.ReadAll(response.Body)

	var resp OrderResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	if resp.Status != "SUCCESS" {
		return nil, errors.New(resp.Message)
	}

	var totlaPrice float64
	for _, ticket := range tickets {
		// TODO IMPLEMENT
		totlaPrice += ticket.TicketType.Price
	}

	var order models.Order = models.Order{
		OrderID:         resp.Data.OrderID,
		MerchantID:      merchant.MerchantID,
		EventID:         uint(terminal.EventID),
		UserUGKthID:     user.UGKthID,
		PaymentPageLink: resp.Data.PaymentPageLink,
		Details: models.OrderDetails{
			OrderID:       resp.Data.OrderID,
			PaymentStatus: models.OrderStatusPending,
			Currency:      "SEK", // TODO: Get currency from terminal
			Total:         totlaPrice,
		},
	}

	err = tx.Session(&gorm.Session{FullSaveAssociations: true}).Create(&order).Error
	if err != nil {
		return nil, err
	}

	return &order, nil
}

func (sos *SurfboardCreateOrderService) generateOrderLines(tickets []models.Ticket) []OrderLine {
	orderLines := []OrderLine{}

	for _, ticket := range tickets {
		ticketType := ticket.TicketType
		orderLine := OrderLine{
			ID:       fmt.Sprintf("ticket-%d", ticket.ID),
			Name:     ticketType.Name,
			Quantity: 1,
			ItemAmount: ItemAmount{
				Regular:  int(ticketType.Price * 100),
				Total:    int(ticketType.Price * 100),
				Currency: "SEK",
				Tax: []Tax{
					{
						Amount:     0, // TODO calculate tax
						Percentage: 0, // TODO calculate tax
						Type:       "vat",
					},
				},
			},
		}
		orderLines = append(orderLines, orderLine)
	}
	return orderLines
}

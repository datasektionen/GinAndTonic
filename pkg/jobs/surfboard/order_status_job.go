package surfboard_job

import (
	"fmt"
	"strings"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	surfboard_service_order "github.com/DowLucas/gin-ticket-release/pkg/services/surfboard/order"
	"gorm.io/gorm"
)

func CheckOrderStatusesJob(db *gorm.DB) {
	// Get all merchants
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	orders, err := models.GetAllIncompleteOrders(tx)
	if err != nil {
		tx.Rollback()
		return
	}
	fmt.Println("Orders: ", len(orders))

	service := surfboard_service_order.NewOrderService()

	for _, order := range orders {
		if order.IsPaymentCompleted() {
			continue
		}

		// Update statuses
		status, err := service.GetOrderStatus(&order)
		if err != nil {
			tx.Rollback()
			return
		}

		fmt.Println("Order status: ", *status)

		statusLowerCase := models.OrderStatus(strings.ToLower(*status))

		if statusLowerCase != order.Status {
			order.Status = statusLowerCase
			if err := tx.Save(&order).Error; err != nil {
				tx.Rollback()
				return
			}
		}
	}

	tx.Commit()
}

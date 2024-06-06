package jobs

import (
	"math"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"gorm.io/gorm"
)

func StartEventSiteVisitsJob(db *gorm.DB) error {
	futureEvents, err := models.GetFutureEvents(db)

	if err != nil {
		return err
	}

	for _, event := range futureEvents {
		eventSiteVisits, err := models.GetEventSiteVisits(db, event.ID)

		if err != nil {
			return err
		}

		ticketOrders, err := models.GetTicketOrdersToEvent(db, event.ID)
		if err != nil {
			return err
		}

		var numTickets int = 0
		for _, ticketOrder := range ticketOrders {
			numTickets += ticketOrder.NumTickets
		}

		totalIncome, err := models.GetEventTotalIncome(db, int(event.ID))
		if err != nil {
			return err
		}

		// Round totalIncome to 2 decimal places
		totalIncome = math.Round(totalIncome / 100)

		err = process_sesvj(db, event.ID, eventSiteVisits, numTickets, totalIncome)

		if err != nil {
			return err
		}
	}

	return nil
}

func process_sesvj(db *gorm.DB, eventID uint, eventSiteVisits []models.EventSiteVisit,
	numTickets int, totalIncome float64) error {
	if len(eventSiteVisits) == 0 {
		return nil
	}

	var summary models.EventSiteVisitSummary = models.EventSiteVisitSummary{
		EventID: eventID,
	}

	var uniqueUsers map[string]bool = make(map[string]bool)

	for _, visit := range eventSiteVisits {
		summary.TotalVisits++

		if _, visited := uniqueUsers[visit.Location]; !visited {
			summary.UniqueUsers++
			uniqueUsers[visit.Location] = true
		}
	}

	summary.NumTickets = numTickets
	summary.TotalIncome = totalIncome

	if err := db.Save(&summary).Error; err != nil {
		return err
	}

	return nil
}

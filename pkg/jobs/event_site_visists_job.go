package jobs

import (
	"math"
	"sync"

	"github.com/DowLucas/gin-ticket-release/pkg/models"
	"gorm.io/gorm"
)

func StartEventSiteVisitsJob(db *gorm.DB) error {
	futureEvents, err := models.GetFutureEvents(db)

	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errs := make(chan error)

	for _, event := range futureEvents {
		wg.Add(1)

		go func(event models.Event) {
			defer wg.Done()

			eventSiteVisits, err := models.GetEventSiteVisits(db, event.ID)

			if err != nil {
				errs <- err
				return
			}

			ticketRequests, err := models.GetTicketRequestsToEvent(db, event.ID)

			if err != nil {
				errs <- err
				return
			}

			totalIncome, err := models.GetEventTotalIncome(db, int(event.ID))
			if err != nil {
				errs <- err
				return
			}

			// Round totalIncome to 2 decimal places
			totalIncome = math.Round(totalIncome / 100)

			err = process_sesvj(db, event.ID, eventSiteVisits, len(ticketRequests), totalIncome)

			if err != nil {
				errs <- err
				return
			}
		}(event)
	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	for err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func process_sesvj(db *gorm.DB, eventID uint, eventSiteVisits []models.EventSiteVisit,
	numTicketReleases int, totalIncome float64) error {
	if len(eventSiteVisits) == 0 {
		return nil
	}

	var summary models.EventSiteVisitSummary = models.EventSiteVisitSummary{
		EventID: eventID,
	}

	var uniqueUsers map[string]bool = make(map[string]bool)

	for _, visit := range eventSiteVisits {
		summary.TotalVisits++

		if _, visited := uniqueUsers[visit.UserUGKthID]; !visited {
			summary.UniqueUsers++
			uniqueUsers[visit.UserUGKthID] = true
		}
	}

	summary.NumTicketRequests = numTicketReleases
	summary.TotalIncome = totalIncome

	if err := db.Save(&summary).Error; err != nil {
		return err
	}

	return nil
}

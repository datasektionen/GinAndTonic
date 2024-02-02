package utils

import (
	"log"
	"time"
)

func ConvertUNIXTimeToDateTime(unixTime int64) time.Time {
	return time.Unix(unixTime, 0)
}

func MustPayBefore(payWithin int, ticketUpdatedAt time.Time) time.Time {
	location, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		log.Fatalf("Failed to load location: %v", err)
	} // TODO Continue here

	ticketUpdatedAt = ticketUpdatedAt.In(location)

	roundedTime := ticketUpdatedAt.Add(time.Duration(payWithin+1) * time.Hour)
	roundedTime = roundedTime.Truncate(time.Hour)
	if roundedTime.Before(ticketUpdatedAt) {
		roundedTime = roundedTime.Add(time.Hour)
	}
	return roundedTime
}

func ConvertPayWithinToString(payWithin int, ticketUpdatedAt time.Time) string {
	mustPayBefore := MustPayBefore(payWithin, ticketUpdatedAt)
	return mustPayBefore.Format("2006-01-02 15:04:05")
}

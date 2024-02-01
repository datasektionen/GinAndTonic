package utils

import "time"

func ConvertUNIXTimeToDateTime(unixTime int64) time.Time {
	return time.Unix(unixTime, 0)
}

func ConvertPayWithinToString(payWithin int, ticketUpdatedAt time.Time) string {
	roundedTime := ticketUpdatedAt.Add(time.Duration(payWithin) * time.Hour)
	roundedTime = roundedTime.Truncate(time.Hour)
	if roundedTime.Before(ticketUpdatedAt) {
		roundedTime = roundedTime.Add(time.Hour)
	}
	return roundedTime.Format("2006-01-02 15:04:05")
}

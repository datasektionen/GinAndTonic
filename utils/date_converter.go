package utils

import "time"

func ConvertUNIXTimeToDateTime(unixTime int64) time.Time {
	return time.Unix(unixTime, 0)
}

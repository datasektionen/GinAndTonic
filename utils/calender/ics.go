package calender_utils

import (
	"fmt"
	"time"
)

func CreateICSLink(summary, description, location, startTime, endTime string) string {
	ics := fmt.Sprintf(`BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//Your Organization//Your Product//EN
BEGIN:VEVENT
UID:%s
DTSTAMP:%s
DTSTART:%s
DTEND:%s
SUMMARY:%s
DESCRIPTION:%s
LOCATION:%s
END:VEVENT
END:VCALENDAR`,
		"unique-id@example.com", // Unique ID for the event
		time.Now().Format("20060102T150405Z"),
		startTime,
		endTime,
		summary,
		description,
		location,
	)

	return ics
}

package calender_utils

import (
	"fmt"
	"net/url"
)

func CreateGoogleCalendarLink(title, description, location, startTime, endTime, email string) string {
	baseURL := "https://calendar.google.com/calendar/render"
	params := url.Values{}
	params.Add("action", "TEMPLATE")
	params.Add("text", title)
	params.Add("details", description)
	params.Add("location", location)
	params.Add("dates", fmt.Sprintf("%s/%s", startTime, endTime))
	params.Add("add", email)

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

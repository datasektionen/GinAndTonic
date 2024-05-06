package utils

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
)

type QueryParams struct {
	Page    int
	PerPage int
	Sort    string
	Order   string
	Filter  string // You might want to use a more structured type depending on your filtering needs
	Range   []int  // New field for the range
}

// GetQueryParams extracts common query parameters from a Gin context
func GetQueryParams(c *gin.Context) (*QueryParams, error) {
	// Default values
	defaultPage := 1
	defaultPerPage := 10
	defaultSort := "id"
	defaultOrder := "ASC"
	defaultRange := []int{0, 24} // Default range

	// Extract pagination parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(defaultPage)))
	if err != nil {
		return nil, err
	}

	perPage, err := strconv.Atoi(c.DefaultQuery("perPage", strconv.Itoa(defaultPerPage)))
	if err != nil {
		return nil, err
	}

	// Extract sorting parameters
	sort := c.DefaultQuery("sort", defaultSort)
	order := c.DefaultQuery("order", defaultOrder)

	// Extract filter parameter
	filter := c.Query("filter") // Assuming a raw JSON string; consider parsing into a structured object

	// Extract range parameter
	rangeStr := c.DefaultQuery("range", "")
	var rangeInt []int
	if rangeStr == "" {
		rangeInt = defaultRange
	} else {
		err := json.Unmarshal([]byte(rangeStr), &rangeInt)
		if err != nil {
			return nil, err
		}
	}

	return &QueryParams{
		Page:    page,
		PerPage: perPage,
		Sort:    sort,
		Order:   order,
		Filter:  filter,
		Range:   rangeInt, // Include the range in the returned QueryParams
	}, nil
}

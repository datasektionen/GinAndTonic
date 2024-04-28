package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type QueryParams struct {
	Page    int
	PerPage int
	Sort    string
	Order   string
	Filter  string // You might want to use a more structured type depending on your filtering needs
}

// GetQueryParams extracts common query parameters from a Gin context
func GetQueryParams(c *gin.Context) (*QueryParams, error) {
	// Default values
	defaultPage := 1
	defaultPerPage := 10
	defaultSort := "id"
	defaultOrder := "ASC"

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

	return &QueryParams{
		Page:    page,
		PerPage: perPage,
		Sort:    sort,
		Order:   order,
		Filter:  filter,
	}, nil
}

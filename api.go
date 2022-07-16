// Implementation of https://github.secureserver.net/CTO/guidelines/tree/master/api-design#pagination
package poctools

import (
	"github.com/gin-gonic/gin"
)

type logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	With(fields map[string]interface{}) logger
}

// This structure comes from the UI with pagination data
type Pagination struct {
	// Maximum number of lines per query, positive means forward values and negative means backward values
	Limit int64
	// A value that will be compared to the database column to get lines bigger that this value
	Marker string
}

type PaginationResponse[T any] struct {
	Data       *[]T                     `json:"data"`
	Pagination PaginationNavigationData `json:"pagination"`
}

type PaginationNavigationData struct {
	First    *string `json:"first"`
	Previous *string `json:"previous"`
	Next     *string `json:"next"`
	Last     *string `json:"last"`
	Total    int64   `json:"total"`
}

// Filter is a generic structure for filtering
type Filter struct {
	// Name is the values expose in the query parameter
	Name string
	// Value is the value of the filter
	Value string
	// Query is the extra 'where clause' that will be added the the principal query to implement this filter
	WhereField string
}

type Order struct {
	// Name is the values expose in the query parameter
	Name string
	// Desc define if it is descendent order or not
	Desc bool
	// Field to be added to the order clause
	OrderField string
}

type ApiParams struct {
	RequestedURLPath string
	Filters          []Filter
	Order            *Order
	Pagination       Pagination
	Options          map[string]bool
	Logger           logger
}

func NoOrders() []Order {
	return make([]Order, 0)
}

func NoFilters() []Filter {
	return make([]Filter, 0)
}

func CreateApiParam(ctx *gin.Context, fields []Filter, orders []Order) ApiParams {

	pagination := GeneratePaginationFromRequest(ctx)
	filters := GenerateFilterFromRequest(ctx, fields)
	order := GenerateOrderFromRequest(ctx, orders)

	requestedURLPath := ctx.Request.URL.Path
	return ApiParams{
		RequestedURLPath: requestedURLPath,
		Filters:          filters,
		Pagination:       pagination,
		Order:            order,
	}
}

func GenerateOrderFromRequest(c *gin.Context, orders []Order) *Order {
	query := c.Request.URL.Query()

	for key, value := range query {
		// Found the order key
		if key == "order" {
			queryValue := value[len(value)-1]
			order := FindOrderByKey(queryValue, orders)
			if order != nil {
				return order
			}
		}

	}
	return nil
}

func GenerateFilterFromRequest(c *gin.Context, fields []Filter) []Filter {
	fs := []Filter{}
	query := c.Request.URL.Query()

	for key, value := range query {
		filter := FindFilterByKey(key, fields)
		if filter != nil {
			queryValue := value[len(value)-1]
			filter.Value = queryValue
			fs = append(fs, *filter)
		}
	}

	return fs
}

func FindFilterByKey(key string, list []Filter) *Filter {
	for i, b := range list {
		if b.Name == key {
			return &list[i]
		}
	}
	return nil
}

func FindOrderByKey(key string, list []Order) *Order {
	for i, b := range list {
		if b.Name == key {
			return &list[i]
		}
	}
	return nil
}

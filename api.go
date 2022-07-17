// Implementation of https://github.secureserver.net/CTO/guidelines/tree/master/api-design#pagination
package poctools

import (
	"github.com/gin-gonic/gin"
)

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
	Logger           interface{}
}

type apiParam struct {
	Options map[string]bool
	logger  interface{}
	ctx     *gin.Context
	filters *[]Filter
	order   *Order
}

func CreateApiParam(ctx *gin.Context) *apiParam {
	return &apiParam{ctx: ctx}
}

func (a *apiParam) WithLog(log interface{}) *apiParam {
	a.logger = log
	return a
}

func (a *apiParam) WithFilters(filters *[]Filter) *apiParam {
	a.filters = filters
	return a
}

func (a *apiParam) WithOrder(order *Order) *apiParam {
	a.order = order
	return a
}

func (a *apiParam) Build() *ApiParams {

	pagination := GeneratePaginationFromRequest(a.ctx)
	var requestOrder *Order
	var requestFilters []Filter

	if a.filters == nil {
		requestFilters = []Filter{}
	} else {
		requestFilters = generateFilterFromRequest(a.ctx, *a.filters)

	}

	if hasOrder(a.ctx, a.order) {
		requestOrder = a.order
	}

	requestedURLPath := a.ctx.Request.URL.Path

	return &ApiParams{
		RequestedURLPath: requestedURLPath,
		Filters:          requestFilters,
		Pagination:       pagination,
		Order:            requestOrder,
		Logger:           a.logger,
	}
}

func hasOrder(c *gin.Context, order *Order) bool {
	query := c.Request.URL.Query()
	for key, value := range query {
		if key == "order" {
			queryValue := value[len(value)-1]
			if order.Name == queryValue {
				return true
			}
		}

	}
	return false
}

func generateFilterFromRequest(c *gin.Context, fields []Filter) []Filter {
	fs := []Filter{}
	query := c.Request.URL.Query()

	for key, value := range query {
		filter := findFilterByKey(key, fields)
		if filter != nil {
			queryValue := value[len(value)-1]
			filter.Value = queryValue
			fs = append(fs, *filter)
		}
	}

	return fs
}

func findFilterByKey(key string, list []Filter) *Filter {
	for i, b := range list {
		if b.Name == key {
			return &list[i]
		}
	}
	return nil
}

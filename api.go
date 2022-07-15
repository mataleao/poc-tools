package poctools

import (
	"github.com/gin-gonic/gin"
	"github.com/mataleao/poctools/dto"
)

func CreateApiParam(ctx *gin.Context, fields []dto.Filter, orders []dto.Order) dto.ApiParams {

	pagination := GeneratePaginationFromRequest(ctx)

	filters := GenerateFilterFromRequest(ctx, fields)
	order := GenerateOrderFromRequest(ctx, orders)

	requestedURLPath := ctx.Request.URL.Path
	return dto.ApiParams{
		RequestedURLPath: requestedURLPath,
		Filters:          filters,
		Pagination:       pagination,
		Order:            order,
	}
}

func GenerateOrderFromRequest(c *gin.Context, orders []dto.Order) *dto.Order {
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

func GenerateFilterFromRequest(c *gin.Context, fields []dto.Filter) []dto.Filter {
	fs := []dto.Filter{}
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

func FindFilterByKey(key string, list []dto.Filter) *dto.Filter {
	for i, b := range list {
		if b.Name == key {
			return &list[i]
		}
	}
	return nil
}

func FindOrderByKey(key string, list []dto.Order) *dto.Order {
	for i, b := range list {
		if b.Name == key {
			return &list[i]
		}
	}
	return nil
}

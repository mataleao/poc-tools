package poctools

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mataleao/poctools/config"
	"github.com/mataleao/poctools/dto"
)

func FindAllPaged[DBE any, DTO any](sql string, s SqlExecutor, params dto.ApiParams, label string, funcMapDbToDto func([]DBE) []DTO, args ...interface{}) (dto.PaginationResponse[DTO], error) {

	response := dto.PaginationResponse[DTO]{}

	if params.Order == nil {
		params.Order = &dto.Order{OrderField: "id"}
	}

	resultList := make([]DBE, 0)
	totalLines, err := s.ReadMany(sql, &resultList, params, args...)
	if err != nil {
		message := fmt.Sprintf("unable to read paged %s", label)
		return response, fmt.Errorf(message)
	}

	// Covert the result for DTOs
	dtos := funcMapDbToDto(resultList)
	response.Data = &dtos

	// Prepare the pagination response
	response.Pagination, err = PreparePaginationResponse(params, totalLines, resultList)
	if err != nil {
		message := fmt.Sprintf("unable to read paged %s", label)
		return response, fmt.Errorf(message)
	}

	return response, err
}

func GeneratePaginationFromRequest(c *gin.Context) dto.Pagination {
	p := dto.Pagination{}

	query := c.Request.URL.Query()
	for key, value := range query {
		queryValue := value[len(value)-1]
		switch key {
		case "limit":
			value, err := strconv.ParseInt(queryValue, 10, 64)
			if err != nil || value == 0 {
				return getDefaultPaginationRequest()
			}
			p.Limit = value
		case "marker":
			p.Marker = queryValue
		}
	}

	if p.Limit == 0 {
		return getDefaultPaginationRequest()
	}

	return p

}

func PreparePaginationResponse(p dto.ApiParams, totalLines int64, list interface{}) (dto.PaginationNavigationData, error) {

	var pnd dto.PaginationNavigationData
	if len(p.RequestedURLPath) == 0 {
		return pnd, fmt.Errorf("the request URL path is empty")
	}

	if totalLines == 0 {
		return pnd, nil
	}

	nextOffset, previousOffset := computeNextAndPrevious(p.Pagination, totalLines)

	pnd, err := getPaginationNavigationData(p.RequestedURLPath, previousOffset, nextOffset, totalLines, p.Pagination.Limit)
	if err != nil {
		return pnd, fmt.Errorf("unable to create pagination data: %w", err)
	}

	return pnd, nil
}

func computeNextAndPrevious(p dto.Pagination, totalLines int64) (int64, int64) {
	marker, err := strconv.ParseInt(p.Marker, 10, 64)
	if err != nil {
		marker = 0
	}

	nextOffset := marker + p.Limit
	previousOffset := marker - p.Limit

	if nextOffset > totalLines {
		nextOffset = totalLines
	}
	if marker == 0 {
		previousOffset = 0
	}

	if p.Marker == "last" {
		previousOffset = totalLines - (p.Limit * 2)
		if previousOffset < 0 {
			previousOffset = 0
		}
		nextOffset = totalLines
	}

	return nextOffset, previousOffset
}

/*
	Fill a PaginationNavigationData instance with values like
	for urls: /v1/users?marker=auser&limit=100

	urlPath the path used in the url. Ex: "/v1/users" "/v1/leads/"
	total 	Total number of lines of the table content
	limit   Limit used to group the total number of lines
*/
func getPaginationNavigationData(urlPath string, previousMarker, nextMarker int64, total, limit int64) (dto.PaginationNavigationData, error) {
	p := dto.PaginationNavigationData{}

	var sb strings.Builder
	sb.WriteString(urlPath)

	if !strings.Contains(urlPath, "?") {
		sb.WriteString("?")
	} else {
		sb.WriteString("&")
	}

	sb.WriteString("marker=%s&limit=%d")

	firstPath := fmt.Sprintf(sb.String(), "", limit)
	p.First = &firstPath
	lastPath := fmt.Sprintf(sb.String(), "last", limit)
	p.Last = &lastPath

	previousPath := fmt.Sprintf(sb.String(), strconv.FormatInt(previousMarker, 10), limit)
	p.Previous = &previousPath

	nextPath := fmt.Sprintf(sb.String(), strconv.FormatInt(nextMarker, 10), limit)
	p.Next = &nextPath

	p.Total = total

	return p, nil
}

func getDefaultPaginationRequest() dto.Pagination {
	return dto.Pagination{Marker: "", Limit: config.DefaultPaginationLimit}
}

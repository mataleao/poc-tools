package poctools

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type paginator[T any] struct {
	s              SqlExecutor
	sql            string
	params         ApiParams
	response       PaginationResponse[T]
	funcMapDbToDto func([]any) []T
	args           []interface{}
}

func PaginatorFor[T any]() *paginator[T] {
	return &paginator[T]{}
}

func (p *paginator[T]) WithSqlExecutor(s SqlExecutor) *paginator[T] {
	p.s = s
	return p
}

func (p *paginator[T]) WithQuery(query string) *paginator[T] {
	p.sql = query
	return p
}

func (p *paginator[T]) WithApiParams(params ApiParams) *paginator[T] {
	p.params = params
	return p
}

func (p *paginator[T]) WithMapperFunc(funcMapDbToDto func([]any) []T) *paginator[T] {
	p.funcMapDbToDto = funcMapDbToDto
	return p
}

func (p *paginator[T]) WithQueryArgs(args ...interface{}) *paginator[T] {
	p.args = append(p.args, args...)
	return p
}

func (p *paginator[T]) Do() (*PaginationResponse[T], error) {

	//Todo test if all fields were populated
	if p.params.Order == nil {
		p.params.Order = &Order{OrderField: "id"}
	}

	var totalLines int64
	var err error

	if p.funcMapDbToDto != nil {
		resultList := make([]interface{}, 0)
		totalLines, err = p.s.readManyPaginated(p.sql, &resultList, p.params, p.args...)
		if err != nil {
			return nil, fmt.Errorf("unable to read paged object")
		}

		dtos := p.funcMapDbToDto(resultList)
		p.response.Data = &dtos
		p.response.Pagination, err = preparePaginationResponse(p.params, totalLines, resultList)
		if err != nil {
			return nil, fmt.Errorf("unable to read paged object")
		}

	} else {
		resultList := make([]T, 0)
		totalLines, err = p.s.readManyPaginated(p.sql, &resultList, p.params, p.args...)
		if err != nil {
			return nil, fmt.Errorf("unable to read paged object")
		}

		p.response.Data = &resultList
		p.response.Pagination, err = preparePaginationResponse(p.params, totalLines, resultList)
		if err != nil {
			return nil, fmt.Errorf("unable to read paged object")
		}
	}

	return &p.response, nil
}

func FindAllPagedMapped[DBE any, DTO any](s SqlExecutor, sql string, params ApiParams, funcMapDbToDto func([]DBE) []DTO, args ...interface{}) (PaginationResponse[DTO], error) {

	response := PaginationResponse[DTO]{}

	if params.Order == nil {
		params.Order = &Order{OrderField: "id"}
	}

	resultList := make([]DBE, 0)
	totalLines, err := s.readManyPaginated(sql, &resultList, params, args...)
	if err != nil {
		return response, fmt.Errorf("unable to read paged object")
	}

	// Covert the result for DTOs
	dtos := funcMapDbToDto(resultList)
	response.Data = &dtos

	// Prepare the pagination response
	response.Pagination, err = preparePaginationResponse(params, totalLines, resultList)
	if err != nil {
		return response, fmt.Errorf("unable to read paged object")
	}

	return response, err
}

func GeneratePaginationFromRequest(c *gin.Context) Pagination {
	p := Pagination{}

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

func preparePaginationResponse(p ApiParams, totalLines int64, list interface{}) (PaginationNavigationData, error) {

	var pnd PaginationNavigationData
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

func computeNextAndPrevious(p Pagination, totalLines int64) (int64, int64) {
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
func getPaginationNavigationData(urlPath string, previousMarker, nextMarker int64, total, limit int64) (PaginationNavigationData, error) {
	p := PaginationNavigationData{}

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

func getDefaultPaginationRequest() Pagination {
	return Pagination{Marker: "", Limit: DefaultPaginationLimit}
}

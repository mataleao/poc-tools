package poctools

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/mataleao/poctools/enum"
)

type SqlExecutor interface {
	ReadMany(sql string, entity interface{}, p ApiParams, pars ...interface{}) (total int64, err error)
	ReadOne(sqlStmt string, entity interface{}, pars ...interface{}) error
	Write(sql string, entity interface{}) (uint64, error)
	GetPaginatedQuery(query string, p ApiParams) (result string, paginationParams []interface{})
	GetFilteredQuery(query string, f []Filter) (result string, pars []interface{})
}

const (
	ROUND_BRACKET_PATTERN = `\([^)()]+\)`
)

type sqlExecutor struct {
	ds DbSession
}

func CreateSqlExecutor(dbSession DbSession) SqlExecutor {
	return &sqlExecutor{ds: dbSession}
}

func (S *sqlExecutor) ReadOne(query string, entity interface{}, pars ...interface{}) error {
	return S.ds.ReadOne(query, entity, pars...)
}

//TODO make apiParam optional
func (S *sqlExecutor) ReadMany(sql string, entity interface{}, apiParam ApiParams, pars ...interface{}) (total int64, err error) {

	queryToBeCounted, extraParsToBeCounted := S.GetFilteredQuery(sql, apiParam.Filters)
	extraParsToBeCounted = append(pars, extraParsToBeCounted...)

	var query string
	query, paginationParams := S.GetPaginatedQuery(queryToBeCounted, apiParam)
	pars = append(extraParsToBeCounted, paginationParams...)

	err = S.ds.ReadMany(query, entity, pars...)
	if err != nil {
		message := "error in paginated query execution"
		return 0, fmt.Errorf(message)
	}

	if apiParam.Pagination.Marker == "last" {
		S.reverseResult(entity)
	}

	if !apiParam.Options[enum.Option.NoCount] {
		var res []int64
		_, hasGroupBy, _, _ := S.testClauses(query)
		var countQuery string

		if hasGroupBy {
			countQuery = fmt.Sprintf("select count(1) from (%s) as cnt", queryToBeCounted)

		} else {
			countQuery = S.replaceInMainQuery(queryToBeCounted, "select.* from", "select count(1) from")
		}
		err = S.ds.ReadMany(countQuery, &res, extraParsToBeCounted...)
		if err != nil {
			message := "error reading total from paginated query execution"
			return 0, fmt.Errorf(message)
		}

		total = res[0]
	}

	return total, err
}

func (*sqlExecutor) reverseResult(entity interface{}) {
	sourceArrPtr := reflect.ValueOf(entity)
	srcArr := reflect.Indirect(sourceArrPtr)

	for i, j := 0, srcArr.Len()-1; i < j; i, j = i+1, j-1 {
		lineA := srcArr.Index(i)
		lineB := srcArr.Index(j)

		tmp := lineA.Interface()
		lineA.Set(lineB)
		lineB.Set(reflect.ValueOf(tmp))
	}
}

func (S *sqlExecutor) Write(sql string, entity interface{}) (uint64, error) {
	return S.ds.Write(sql, entity)
}

func (S *sqlExecutor) GetPaginatedQuery(query string, p ApiParams) (result string, paginationParams []interface{}) {

	ordered := false
	marker, err := strconv.Atoi(p.Pagination.Marker)
	if err != nil {
		if len(p.Pagination.Marker) == 0 {
			marker = 0
		}

		if p.Pagination.Marker == "last" {
			_, _, _, hasDesc := S.testClauses(query)

			if hasDesc {
				tokens := S.queryTokens(query)
				tokens = tokens[:len(tokens)-1] // remove last token tha must be desc
				query = strings.Join(tokens, " ")

			} else {

				if p.Order != nil {
					query = fmt.Sprintf("%s order by %s desc", query, p.Order.OrderField)
				} else {
					query = fmt.Sprintf("%s order by id desc", query)
				}

				ordered = true
			}

			marker = 0
		}

	}

	paginationParams = append(paginationParams, p.Pagination.Limit, marker)

	if !ordered && p.Order != nil {
		if p.Order.Desc {
			query = fmt.Sprintf("%s order by %s desc", query, p.Order.OrderField)
		} else {
			query = fmt.Sprintf("%s order by %s", query, p.Order.OrderField)
		}
	}

	query = fmt.Sprintf("%s limit ? offset ?", query)
	query = strings.Trim(query, " ")
	return query, paginationParams
}

func (S *sqlExecutor) GetFilteredQuery(query string, f []Filter) (result string, pars []interface{}) {

	if len(f) == 0 {
		return query, pars
	}

	hasWhere, hasGroupBy, _, _ := S.testClauses(query)
	if !hasWhere {
		if hasGroupBy {
			var filterConditions string
			pars, filterConditions = S.appendFiltersConditions(f, pars, "")
			replaceFor := fmt.Sprintf(" where %s group by ", filterConditions)
			query = S.replaceInMainQuery(query, " group by ", replaceFor)

		} else {
			query = fmt.Sprintf("%s where ", query)
			pars, query = S.appendFiltersConditions(f, pars, query)
		}
	} else {
		if hasGroupBy {
			var filterConditions string
			pars, filterConditions = S.appendFiltersConditions(f, pars, "")
			replaceFor := fmt.Sprintf(" %s and ", filterConditions)
			query = S.replaceInMainQuery(query, " where ", replaceFor)

		} else {
			query = fmt.Sprintf("%s and ", query)
			pars, query = S.appendFiltersConditions(f, pars, query)
		}
	}
	query = strings.Trim(query, " ")
	return query, pars
}

func (*sqlExecutor) appendFiltersConditions(f []Filter, pars []interface{}, query string) ([]interface{}, string) {
	for _, filter := range f {
		pars = append(pars, filter.Value)
		query = fmt.Sprintf("%s%s=? ", query, filter.WhereField)
	}
	query = strings.Trim(query, " ")
	return pars, query
}

// testClauses compute if the given query has where and order by clause
func (S *sqlExecutor) testClauses(query string) (hasWhere, hasGroupBy, hasOrderBy, hasDesc bool) {
	query = S.removeAllSubQueries(query)
	query = RemoveQueryLinters(query)
	hasWhere = strings.Index(query, " where ") > 0
	hasGroupBy = strings.Index(query, " group by ") > 0
	hasOrderBy = strings.Index(query, " order by ") > 0

	tokens := S.queryTokens(query)
	lastToken := tokens[len(tokens)-1]
	hasDesc = lastToken == "desc"

	return hasWhere, hasGroupBy, hasOrderBy, hasDesc
}

func (*sqlExecutor) queryTokens(query string) []string {
	return strings.Split(query, " ")
}

// removeAllSubQueries only round brackets from the given query as it is a pattern of sub query
func (S *sqlExecutor) removeAllSubQueries(query string) string {
	sampleRegexp := regexp.MustCompile(ROUND_BRACKET_PATTERN)
	result := sampleRegexp.FindString(query)

	for len(result) > 0 {
		query = strings.Replace(query, result, "SQ", 1)
		result = sampleRegexp.FindString(query)
	}

	return query
}

// replaceInMainQuery replaces some pattern from the main query
//                    the main query is the given minus all supposed sub queries
func (S *sqlExecutor) replaceInMainQuery(query, replace, newValue string) string {
	regex := regexp.MustCompile(ROUND_BRACKET_PATTERN)
	result := regex.FindString(query)
	var subs []string
	idx := 0

	for len(result) > 0 {
		subs = append(subs, result)
		subQueryLabel := fmt.Sprintf("_sq_%d_", idx)
		idx++
		query = strings.Replace(query, result, subQueryLabel, 1)
		result = regex.FindString(query)
	}

	query = RemoveQueryLinters(query)
	replace = strings.ToLower(replace)
	newValue = strings.ToLower(newValue)

	regex = regexp.MustCompile(replace)
	query = regex.ReplaceAllString(query, newValue)

	for idx := range subs {
		k := len(subs) - 1 - idx // now k starts from the end
		sub := subs[k]
		subQueryLabel := fmt.Sprintf("_sq_%d_", k)

		query = strings.Replace(query, subQueryLabel, sub, 1)
	}

	return query
}

// RemoveQueryLinters just lower and remove query linters
func RemoveQueryLinters(query string) string {
	query = strings.ToLower(query)
	regex := regexp.MustCompile(`[\n\t]+`)
	query = regex.ReplaceAllString(query, " ")

	regex = regexp.MustCompile(`[ ]+`)
	query = regex.ReplaceAllString(query, " ")
	return query
}

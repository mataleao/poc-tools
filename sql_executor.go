package poctools

import (
	"fmt"
)

type SqlExecutor interface {
	ReadMany(sql string, entity interface{}, pars ...interface{}) error
	readManyPaginated(sql string, entity interface{}, p ApiParams, pars ...interface{}) (total int64, err error)
	ReadOne(sqlStmt string, entity interface{}, pars ...interface{}) error
	Write(sql string, entity interface{}) (uint64, error)
}

type sqlExecutor struct {
	ds DbSession
}

func CreateSqlExecutor(dbSession DbSession) SqlExecutor {
	return &sqlExecutor{ds: dbSession}
}

func (S *sqlExecutor) ReadOne(query string, entity interface{}, pars ...interface{}) error {
	return S.ds.ReadOne(query, entity, pars...)
}

func (S *sqlExecutor) ReadMany(query string, entity interface{}, pars ...interface{}) error {

	err := S.ds.ReadMany(query, entity, pars...)
	if err != nil {
		message := "error in query execution"
		return fmt.Errorf(message)
	}

	return nil
}

func (S *sqlExecutor) readManyPaginated(sql string, entity interface{}, apiParam ApiParams, pars ...interface{}) (total int64, err error) {

	queryToBeCounted, extraParsToBeCounted := getFilteredQuery(sql, apiParam.Filters)
	extraParsToBeCounted = append(pars, extraParsToBeCounted...)

	var query string
	query, paginationParams := getPaginatedQuery(queryToBeCounted, apiParam)
	pars = append(extraParsToBeCounted, paginationParams...)

	err = S.ds.ReadMany(query, entity, pars...)
	if err != nil {
		message := "error in paginated query execution"
		return 0, fmt.Errorf(message)
	}

	if apiParam.Pagination.Marker == "last" {
		S.reverseResult(entity)
	}

	if !apiParam.Options[Option.NoCount] {
		var res []int64
		_, hasGroupBy, _, _ := testClauses(query)
		var countQuery string

		if hasGroupBy {
			countQuery = fmt.Sprintf("select count(1) from (%s) as cnt", queryToBeCounted)

		} else {
			countQuery = replaceInMainQuery(queryToBeCounted, "select.* from", "select count(1) from")
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

func (S *sqlExecutor) Write(sql string, entity interface{}) (uint64, error) {
	return S.ds.Write(sql, entity)
}

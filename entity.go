package poctools

import (
	"fmt"
	"strings"
	"time"
)

type Entity struct {
	Id        uint64    `db:"id"`
	CreatedAt time.Time `db:"created_at"`
}

func GetBaseFields() []string {
	return []string{
		"id",
		"created_at",
	}
}

type IEntity interface {
	GetId() uint64
	GetTableName() string
	GetFields() []string
}

func GetQuery(e IEntity, filters ...string) string {
	sqlStmt := fmt.Sprintf("select %s, %s from %s",
		strings.Join(GetBaseFields(), ", "), strings.Join(e.GetFields(), ", "), e.GetTableName())

	return sqlPrepareWhere(sqlStmt, filters...)
}

func SaveById(e IEntity) string {

	fields := removeForDML(e.GetFields())

	if e.GetId() > 0 {
		pattern := "%s%s=:%s%s"
		updateField := pattern
		comma := ""

		for _, f := range fields {
			updateField = fmt.Sprintf(updateField, comma, f, f, pattern)
			comma = ", "
		}

		updateField = strings.Replace(updateField, pattern, "", 1)

		return fmt.Sprintf("update %s set %s where id=:id", e.GetTableName(), updateField)
	}

	return fmt.Sprintf("insert into %s (%s) values (:%s)",
		e.GetTableName(),
		strings.Join(fields, ", "),
		strings.Join(fields, ", :"))
}

func removeForDML(s []string) []string {
	var r []string
	for i, v := range s {
		if v != "created_at" && v != "updated_at" {
			r = append(r, s[i])
		}
	}
	return r
}

func sqlPrepareWhere(sqlStmt string, filters ...string) string {
	if len(filters) == 0 {
		return sqlStmt
	}
	filterStr := strings.Join(filters, " and ")

	return fmt.Sprintf("%s where %s", sqlStmt, filterStr)

}

package sqlexecutor

import (
	"github.com/jmoiron/sqlx"
)

var databaseEngine *sqlx.DB

func GetDbEngine() *sqlx.DB {
	return databaseEngine
}

func SetDbEngine(e *sqlx.DB) {
	databaseEngine = e
}

func GetTransactionObject() (*sqlx.Tx, error) {
	return GetDbEngine().Beginx()
}

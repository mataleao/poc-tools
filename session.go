package poctools

import (
	"github.com/jmoiron/sqlx"
)

type DbSession interface {
	ReadOne(sqlStmt string, entity interface{}, pars ...interface{}) error
	ReadMany(query string, entity interface{}, pars ...interface{}) error
	Write(query string, entity interface{}) (uint64, error)
	Close(aborted bool) error
	SetAutoCommit(auto bool)
}

var dbSessionMock DbSession

func MockDbSession(m DbSession) {
	dbSessionMock = m
}

func DbSessionCreate(autoCommit bool) DbSession {
	if dbSessionMock != nil {
		return dbSessionMock
	}

	return &dbSessionImpl{autoCommit: autoCommit}
}

type dbSessionImpl struct {
	tx         *sqlx.Tx
	autoCommit bool
}

func (S *dbSessionImpl) ReadOne(query string, entity interface{}, pars ...interface{}) error {
	if S.tx != nil {
		return S.tx.QueryRowx(query, pars...).StructScan(entity)
	}
	return GetDbEngine().QueryRowx(query, pars...).StructScan(entity)
}

func (S *dbSessionImpl) ReadMany(query string, entity interface{}, pars ...interface{}) error {
	if S.tx != nil {
		return S.tx.Select(entity, query, pars...)
	}

	return GetDbEngine().Select(entity, query, pars...)
}

func (S *dbSessionImpl) Write(sql string, entity interface{}) (uint64, error) {
	var err error

	if S.tx == nil {
		S.tx, err = GetTransactionObject()
		if err != nil {
			return 0, err
		}
	}

	result, err := S.tx.NamedExec(sql, entity)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if S.autoCommit {
		err = S.commit()
		if err != nil {
			err = S.rollback()
		}
		S.tx = nil
	}

	return uint64(id), err
}

func (S *dbSessionImpl) Close(aborted bool) error {
	if S.tx == nil {
		return nil
	}

	if aborted {
		return S.rollback()
	}

	return S.commit()
}

func (S *dbSessionImpl) SetAutoCommit(auto bool) {
	S.autoCommit = auto
}

func (S *dbSessionImpl) commit() error {
	err := S.tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (S *dbSessionImpl) rollback() error {
	err := S.tx.Rollback()
	if err != nil {
		return err
	}
	return nil
}

package database

import (
	"database/sql"
)

type UserQuery struct {
	db *sql.DB
}

var defaultUserQuery *UserQuery = NewUserQuery()

func NewUserQuery() *UserQuery {
	return &UserQuery{
		db: NewMysql(),
	}
}

package database

import (
	"fmt"
)

// func (u *UserQuery) CompareMD5Passwd(passwd string) bool {

// }

// func (u *UserQuery) AddUser() {

// }

// func (u *UserQuery) DeleteUser() {

// }

// func (u *UserQuery) UpdateUser(uid uint64) bool {

// }

func HasUser(uid uint64) (bool, error) {
	if defaultUserQuery == nil {
		return false, fmt.Errorf("defaultUserQuery is nil")
	}
	return defaultUserQuery.HasUser(uid)
}

func (u *UserQuery) HasUser(uid uint64) (bool, error) {
	if u.db == nil {
		return false, fmt.Errorf("database is nil")
	}
	db := u.db
	defer db.Close()
	rows, err := db.Query("select uid from user where uid = ?", uid)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	if rows.Next() {
		return true, nil
	}
	return false, nil
}

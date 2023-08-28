package models

import (
	"log"
	"time"

	"github.com/ciiim/cloudborad/errmsg"
	"github.com/ciiim/cloudborad/internal/database"
)

type User struct {
	UID             int64     `json:"uid"`
	NickName        string    `json:"nickname"`
	Passwd          string    `json:"passwd"`
	PermissionGroup int       `json:"permission_group"`
	Sex             int       `json:"sex"`
	Email           string    `json:"email"`
	PhoneNumber     string    `json:"phone_number"`
	RegisterTime    time.Time `json:"register_time"`
	LastLoginTime   time.Time `json:"lastlogin_time"`
	AccountStatus   int       `json:"account_status"`
}

type UserBasic struct {
	UID             int64  `json:"uid"`
	Email           string `json:"email"`
	NickName        string `json:"nickname"`
	Passwd          string `json:"passwd"`
	PermissionGroup int    `json:"permission_group"`
}

type NewUser struct {
	NickName      string    `json:"nickname"`
	Email         string    `json:"email"`
	Passwd        string    `json:"passwd"`
	PhoneNumber   string    `json:"phone_number"`
	RegisterTime  time.Time `json:"register_time"`
	AccountStatus int       `json:"account_status"`
}

const UserTable = "users"

func QueryUserBasic(email string) (UserBasic, error) {
	stmt, err := database.MysqlDB.Prepare("SELECT uid, email, nickname, passwd, permission_group FROM users WHERE email = ?")
	if err != nil {
		log.Printf("[QueryUserBasic] Prepare failed: %v", err)
		return UserBasic{}, errmsg.ErrDatabaseInternalError
	}
	rows, err := stmt.Query(email)
	if err != nil {
		log.Printf("[QueryUserBasic] Query failed: %v", err)
		return UserBasic{}, errmsg.ErrQueryUserFailed
	}
	for rows.Next() {
		var user UserBasic
		if err := rows.Scan(&user.UID, &user.Email, &user.NickName, &user.Passwd, &user.PermissionGroup); err == nil {
			return user, nil
		}
	}
	return UserBasic{}, errmsg.ErrUserNotFound
}

func InsertUserBasic(newUser *NewUser) (uid int64, err error) {
	stmt, err := database.MysqlDB.Prepare("INSERT INTO users(email, nickname, passwd, phone_number, register_time ,account_status) VALUES(?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Printf("[InsertUserBasic] Prepare failed: %v", err)
		return -1, errmsg.ErrDatabaseInternalError
	}
	res, err := stmt.Exec(newUser.Email, newUser.NickName, newUser.Passwd, newUser.PhoneNumber, newUser.RegisterTime, newUser.AccountStatus)
	if err != nil {
		log.Printf("[InsertUserBasic] Exec failed: %v", err)
		return -1, errmsg.ErrInsertUserFailed
	}
	uid, err = res.LastInsertId()
	return
}

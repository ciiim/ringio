package models

import (
	"errors"
	"log"
	"time"

	"github.com/ciiim/cloudborad/errmsg"
	"github.com/ciiim/cloudborad/internal/database"
	"github.com/redis/go-redis/v9"
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

func EmailExist(email string) (bool, error) {
	stmt, err := database.MysqlDB.Prepare("SELECT uid FROM users WHERE email = ?")
	if err != nil {
		log.Printf("[EmailExist] Prepare failed: %v", err)
		return false, errmsg.ErrDatabaseInternalError
	}
	rows, err := stmt.Query(email)
	if err != nil {
		log.Printf("[EmailExist] Query failed: %v", err)
		return false, errmsg.ErrQueryUserFailed
	}
	for rows.Next() {
		return true, nil
	}
	return false, nil
}

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

func UpdateUserPasswd(email, passwd string) error {
	stmt, err := database.MysqlDB.Prepare("UPDATE users SET passwd = ? WHERE email = ?")
	if err != nil {
		log.Printf("[UpdateUserPasswd] Prepare failed: %v", err)
		return errmsg.ErrDatabaseInternalError
	}
	_, err = stmt.Exec(passwd, email)
	if err != nil {
		log.Printf("[UpdateUserPasswd] Exec failed: %v", err)
		return errmsg.ErrUpdateUserFailed
	}
	return nil
}

func SetVerifyCode(email, codeType, code string, expireTime time.Duration) error {
	err := database.RedisSet(email+codeType, code, expireTime)
	if err != nil {
		return err
	}
	return nil
}

func GetVerifyCode(email string) (string, error) {
	res := database.RedisGet(email)
	if res.Err() != nil && !errors.Is(res.Err(), redis.Nil) {
		return "", res.Err()
	}
	return res.Val(), nil
}

func DeleteVerifyCode(email string) error {
	err := database.RedisDel(email)
	if err != nil {
		return err
	}
	return nil
}

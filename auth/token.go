package auth

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ciiim/cloudborad/auth/cipher"
)

const key = "Sf5asj2k$n13938&"

// const tokenSuffix = "token"

var en cipher.Cipher = cipher.NewAES(key)

type token struct {
	token      string
	expireTime time.Time
}

func (t token) GetToken() string {
	return t.token
}

// func (t token) GetUID() uint64 {
// 	data, err := en.Decrypt([]byte(t.token))
// 	if err != nil {
// 		return 0
// 	}
// 	uid, _, err := getData(data)
// 	if err != nil {
// 		return 0
// 	}
// 	return uid
// }

func NewToken(uid uint64) *token {
	expireTime := time.Now().Add(time.Hour * 720)
	plainText := fmt.Sprintf("%d|%d", uid, expireTime.UnixMilli())
	t, err := en.Encrypt([]byte(plainText))
	if err != nil {
		return &token{}
	}
	return &token{
		token:      string(t),
		expireTime: expireTime,
	}
}

func getData(rawData []byte) (uint64, int64, error) {
	rawDataStr := string(rawData)
	data := strings.SplitN(rawDataStr, "|", 2)
	if len(data) != 2 {
		return 0, 0, fmt.Errorf("decrypt data len less than 2")
	}
	uid, err := strconv.ParseUint(data[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	timestamp, err := strconv.ParseInt(data[1], 10, 64)
	if err != nil {
		return uid, 0, err
	}

	return uid, timestamp, nil
}

func (t *token) Refresh(plainText string) error {
	newToken, err := en.Encrypt([]byte(plainText))
	if err != nil {
		return err
	}
	t.token = string(newToken)
	t.expireTime = time.Now()
	return nil
}

func (t *token) Check() (uint64, IdentifyState, error) {
	bytes, _ := en.Decrypt([]byte(t.token))
	uid, timestamp, err := getData(bytes)
	if uid == 0 {
		return uid, Invaild, fmt.Errorf("invaild uid:%d", uid)
	}
	if timestamp < time.Now().UnixMilli() {
		return uid, Expired, fmt.Errorf("expired")
	}
	if err != nil {
		return uid, Invaild, err
	}
	return uid, Vaild, nil
}

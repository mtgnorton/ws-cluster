package auth

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/crypto/gaes"
)

var ErrFormat = fmt.Errorf("format error")
var DefaultKey = "1234567890123456"

type UserData struct {
	PID        string
	UID        string
	ClientType int // client.CType
}

func Decode(s string, key ...string) (*UserData, error) {
	aesKey := DefaultKey
	if len(key) > 0 {
		aesKey = key[0]
	}

	// 判断长度是否是3的倍数，如果不是增加=
	if len(s)%3 != 0 {
		s += strings.Repeat("=", 3-len(s)%3)
	}

	base64, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	r, err := gaes.Decrypt(base64, []byte(aesKey))
	if err != nil {
		return nil, err
	}

	d := string(r)
	fmt.Println(d)
	arr := strings.Split(d, "@")
	if len(arr) != 3 {
		return nil, ErrFormat
	}
	userData := UserData{
		PID: arr[0],
		UID: arr[1],
	}
	if c, err := strconv.Atoi(arr[2]); err != nil {
		return nil, err
	} else {
		userData.ClientType = c
	}

	return &userData, nil
}

func Encode(userData *UserData, key ...string) (string, error) {
	aesKey := DefaultKey
	if len(key) > 0 {
		aesKey = key[0]
	}
	s := fmt.Sprintf("%s@%s@%d", userData.PID, userData.UID, userData.ClientType)
	c, err := gaes.Encrypt([]byte(s), []byte(aesKey))
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(c), nil
}

func MustEncode(userData *UserData, key ...string) string {
	s, err := Encode(userData, key...)
	if err != nil {
		panic(err)
	}
	return s
}

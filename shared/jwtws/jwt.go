package jwtws

import (
	"time"

	"ws-cluster/config"

	jwt "github.com/dgrijalva/jwt-go"
)

type JwtWs struct {
	secret     []byte
	expireTime int // token过期时间，单位小时
}

type Claims struct {
	PID        string `json:"pid"`
	UID        string `json:"uid"`
	ClientType int    `json:"client_type"`
	jwt.StandardClaims
}

func NewJwtWs(c config.Config) *JwtWs {
	return &JwtWs{
		secret:     []byte(c.Values().Jwt.Secret),
		expireTime: c.Values().Jwt.Expire,
	}
}

// Generate 生成jwt, pid, uid, clientType类型见 client.CType
func (j *JwtWs) Generate(pid, uid string, clientType int) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(time.Hour * time.Duration(j.expireTime))

	var claims = Claims{
		PID:        pid,
		UID:        uid,
		ClientType: clientType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			Issuer:    "ws-cluster",
		},
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(j.secret)

	return token, err
}

func (j *JwtWs) Parse(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return j.secret, nil
	})

	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}

	return nil, err
}

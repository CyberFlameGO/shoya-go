package models

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"gitlab.com/george/shoya-go/config"
	"time"
)

var ErrInvalidAuthCookie = errors.New("invalid auth cookie")

// AuthCookieClaims is the struct that will be encoded to the JWT
type AuthCookieClaims struct {
	UserID      string `json:"uid"`
	IpAddress   string `json:"ip"`
	ClientToken bool   `json:"_c"`
	jwt.StandardClaims
}

// CreateAuthCookie creates a new JWT for the user with the given ID & IP address
func CreateAuthCookie(u *User, ip string, isClientToken bool) (string, error) {
	claims := AuthCookieClaims{
		UserID:      u.ID,
		IpAddress:   ip,
		ClientToken: isClientToken,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.ApiConfiguration.JwtSecret.Get()))
}

// ValidateAuthCookie validates the given JWT and returns the user ID if it is valid
func ValidateAuthCookie(token string, ip string, isClientRequest bool, isPhotonRequest bool) (string, error) {
	claims := AuthCookieClaims{}
	tkn, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.ApiConfiguration.JwtSecret.Get()), nil
	})

	if err != nil {
		return "", err
	}

	if !tkn.Valid {
		return "", ErrInvalidAuthCookie
	}

	if !isPhotonRequest && claims.IpAddress != ip {
		return "", ErrInvalidAuthCookie
	}

	if !isPhotonRequest && claims.ClientToken != isClientRequest {
		return "", ErrInvalidAuthCookie
	}

	return claims.UserID, nil
}

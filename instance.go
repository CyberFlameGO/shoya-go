package main

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"time"
)

var ErrInvalidJoinJWT = errors.New("invalid join token")

type InstanceJoinJWTClaims struct {
	JoinId          string   `json:"joinId"`
	UserId          string   `json:"userId"`
	Session         string   `json:"session"`
	IP              string   `json:"ip"`
	Location        string   `json:"location"`
	WorldAuthorId   string   `json:"worldAuthorId"`
	WorldCapacity   int      `json:"worldCapacity"`
	WorldName       string   `json:"worldName"`
	WorldTags       []string `json:"worldTags"`
	InstanceOwnerId string   `json:"instanceOwnerId"`
	jwt.StandardClaims
}

func CreateJoinToken(u *User, w *World, ip string, location string) (string, error) {
	joinId, _ := uuid.NewUUID()
	claims := InstanceJoinJWTClaims{
		JoinId:          "join_" + joinId.String(),
		UserId:          u.ID,
		Session:         "", // Unknown at the moment.
		IP:              ip,
		Location:        location,
		WorldAuthorId:   w.AuthorID,
		WorldName:       w.Name,
		WorldTags:       w.Tags,
		InstanceOwnerId: "", // TODO: parseLocationString()
		StandardClaims: jwt.StandardClaims{
			Audience:  "VRChatNetworking",
			Issuer:    "VRChat",
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(ApiConfiguration.JwtSecret.Get()))
}

func ValidateJoinToken(token string) (*InstanceJoinJWTClaims, error) {
	claims := InstanceJoinJWTClaims{}
	tkn, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(ApiConfiguration.JwtSecret.Get()), nil
	})

	if err != nil {
		return nil, err
	}

	if !tkn.Valid {
		return nil, ErrInvalidJoinJWT
	}

	return &claims, nil
}

package internal

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type MyCustomClaims struct {
	jwt.RegisteredClaims
}

func AuthorizationHeader(header string) (string, error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", errors.New("unauthorized")
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix)), nil
}

func CreateJwt(user *User, jwtSecret []byte, expiresInSeconds int) (string, error) {
	tokenExpiration := time.Now().Add(24 * time.Hour)
	if expiresInSeconds > 0 {
		tokenExpiration = time.Now().Add(time.Duration(expiresInSeconds * int(time.Second)))
	}
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(tokenExpiration),
		Subject:   fmt.Sprint(user.Id),
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateToken(headerToken string, jwtSecret []byte, claims *MyCustomClaims) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(headerToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("theres a problem with the signing method")
		}
		return jwtSecret, nil
	})
	return token, err
}

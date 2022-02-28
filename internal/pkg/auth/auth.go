package auth

import (
	"errors"
	"fmt"
	"ryzenlo/to2cloud/configs"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var jwtKey string

func init() {
	jwtKey = configs.Conf.JWT.Key
}

func CreateToken(userID int64, isRoot int) (string, error) {
	expiedAt := time.Now().Add(time.Hour * 24 * 7).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"is_root": isRoot,
		"exp":     expiedAt,
	})
	ss, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		err = fmt.Errorf("jwt error:%w", err)
	}
	return ss, err
}

func GetDataFromToken(token string) (map[string]interface{}, error) {
	jwtToken, err := VerifyToken(token)
	if err != nil {
		return nil, err
	}
	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid jwt token")
	}
	return claims, nil
}

func VerifyToken(token string) (*jwt.Token, error) {
	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %w", token.Header["alg"])
		}
		return []byte(jwtKey), nil
	})
	if err != nil {
		return nil, err
	}
	if !jwtToken.Valid {
		return nil, fmt.Errorf("invalid jwt token")
	}
	return jwtToken, nil
}

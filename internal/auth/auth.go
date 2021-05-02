package auth

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"
	"uacl/messages"
	"uacl/model"

	"github.com/TomBowyerResearchProject/common/logger"
	"github.com/dgrijalva/jwt-go"
)

const expirationTime = 100000

func CreateToken(user model.User) (string, error) {
	expiresAt := time.Now().Add(time.Minute * expirationTime).Unix()

	now := time.Now().UTC()

	short := model.ShortenedUser{
		Name:     user.Name,
		Username: user.Username,
	}

	claims := make(jwt.MapClaims)
	claims["dat"] = short
	claims["exp"] = expiresAt
	claims["iat"] = now.Unix()
	claims["nbf"] = now.Unix()

	private, err := ioutil.ReadFile(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		return "", err
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(private)
	if err != nil {
		return "", err
	}

	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
	if err != nil {
		return "", err
	}

	logger.Info("Successfully created token")

	return tokenString, nil
}

func Validate(token string) (model.ShortenedUser, error) {
	var shorten model.ShortenedUser

	public, err := ioutil.ReadFile(os.Getenv("PUBLIC_KEY"))
	if err != nil {
		return shorten, err
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(public)
	if err != nil {
		return shorten, err
	}

	tok, err := jwt.Parse(token, func(jwtToken *jwt.Token) (interface{}, error) {
		if _, ok := jwtToken.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, messages.ErrUnexpectedMethod
		}

		return key, nil
	})
	if err != nil {
		return shorten, err
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok || !tok.Valid {
		return shorten, messages.ErrInvalid
	}

	jsonString, err := json.Marshal(claims["dat"])
	if err != nil {
		return shorten, err
	}

	err = json.Unmarshal(jsonString, &shorten)
	if err != nil {
		return shorten, err
	}

	logger.Info("Successfully validated token")

	return shorten, nil
}

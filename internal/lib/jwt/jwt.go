package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"ssov2/internal/domain/models"
	"time"
)

const secret = "kldjlkadsjfdasjfdhnbxjcnkx"

func NewToken(user models.User, duration time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["uid"] = user.ID
	claims["exp"] = time.Now().Add(duration).Unix()

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

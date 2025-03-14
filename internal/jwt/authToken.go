package authToken

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"go-chat/internal/database"
	"os"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/golang-jwt/jwt/v5"
)

// type of database.User that can be read using jwt token
type JwtUser struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Uid      string `json:"uid"`
}

func CreateAccessToken(user database.User) string {
	return CreateToken(user, time.Now().Add(time.Duration(time.Minute*15)).Unix())
}

func CreateRefreshToken(user database.User) string {
	return CreateToken(user, time.Now().AddDate(0, 1, 0).Unix())
}

func CreateToken(user database.User, expires int64) string {
	authSecret := os.Getenv("authSecret")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":      "auth-server",
		"sub":      user.Username,
		"username": user.Username,
		"email":    user.Email,
		"uid":      user.Uid,
		"exp":      expires,
	})

	signedToken, _ := token.SignedString([]byte(authSecret))

	jwt.WithValidMethods([]string{"HS256"})

	return signedToken
}

func VerifyToken(tokenString string) error {
	authSecret := os.Getenv("authSecret")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(authSecret), nil
	})

	if err != nil {
		return err
	}

	date, err := token.Claims.GetExpirationTime()

	if err != nil {
		return err
	}

	if date == nil || time.Now().Unix() > date.Unix() {
		return fmt.Errorf("token expired")
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	return nil
}

func decodeBase64(encoded string) ([]byte, error) {
	encoded = strings.TrimSpace(encoded)

	if len(encoded)%4 != 0 {
		encoded += strings.Repeat("=", 4-len(encoded)%4)
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return []byte{}, err
	}
	return decodedBytes, nil
}

func GetUserData(tokenString string) (JwtUser, error) {
	sect := strings.Split(tokenString, ".")

	decodedBytes, err := decodeBase64(sect[1])

	if err != nil {
		return JwtUser{}, err
	}

	userData := JwtUser{}

	err = json.Unmarshal(decodedBytes, &userData)

	if err != nil {
		return JwtUser{}, err
	}

	return userData, nil
}

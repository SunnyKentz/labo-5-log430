package authData

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your-very-secret-key") // Replace with env/config in production

// CreateJWT generates a JWT for the given username
func CreateJWT(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateJWT parses and validates the JWT, returning the username if valid
func ValidateJWT(tokenString string) bool {
	// Remove "Bearer " prefix if present
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return false
	}
	if _ /*claims*/, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		//username, ok := claims["username"].(string)
		if !ok {
			return false // errors.New("username not found in token")
		}
		return true
	}
	return false // errors.New("invalid token")
}

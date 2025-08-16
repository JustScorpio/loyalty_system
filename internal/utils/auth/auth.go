package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// Имя куки с JWT-токеном
	JwtCookieName = "jwt_token"
	//Время жизни токена
	TokenLifeTime = time.Hour * 3
	// Ключ для генерации и расшифровки токена (В РЕАЛЬНОМ ПРИЛОЖЕНИИ ХРАНИТЬ В НАДЁЖНОМ МЕСТЕ)
	secretKey = "supersecretkey"
)

// Claims — структура утверждений, которая включает стандартные утверждения и одно пользовательское UserID
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

// newJWTString создаёт токен и возвращает его в виде строки.
func GenerateToken(userID string) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// Срок окончания времени жизни токена
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenLifeTime)),
		},
		// собственное утверждение
		UserID: userID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

// ParseToken проверяет токен и возвращает claims
func ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

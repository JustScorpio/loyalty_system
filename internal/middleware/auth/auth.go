package auth

import (
	"net/http"
	"time"

	"github.com/JustScorpio/loyalty_system/internal/customcontext"
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

// middleware для добавления и чтения кук.
func AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(JwtCookieName)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := ParseToken(cookie.Value)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := customcontext.WithUserID(r.Context(), claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

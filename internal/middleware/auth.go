package middleware

import (
	"net/http"

	"github.com/JustScorpio/loyalty_system/internal/customcontext"
	"github.com/JustScorpio/loyalty_system/internal/utils/auth"
)

// middleware для добавления и чтения кук.
func AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(auth.JwtCookieName)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := auth.ParseToken(cookie.Value)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			ctx := customcontext.WithUserID(r.Context(), claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

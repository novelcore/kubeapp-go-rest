package middleware

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/novelcore/kubeapp-go-rest/internal/auth"
	"github.com/sirupsen/logrus"
)

// JWTAuth validates Bearer tokens and rejects requests without a valid token
func JWTAuth(validator *auth.JWTValidator, log *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			authHeader := r.Header.Get("Authorization")
			userTokenHeader := r.Header.Get("X-User-Token")

			tokenString, err := auth.ExtractToken(authHeader, userTokenHeader)
			if err != nil {
				log.Warn("Missing JWT token in request")
				writeJSONError(w, http.StatusUnauthorized, "missing_token", "Authentication token required")
				return
			}

			authCtx, err := validator.ExtractAuthContext(ctx, tokenString)
			if err != nil {
				errMsg := err.Error()
				if strings.Contains(errMsg, "expired") || strings.Contains(errMsg, "token is expired") {
					log.WithError(err).Warn("JWT token has expired")
					writeJSONError(w, http.StatusUnauthorized, "token_expired", "Authentication token has expired")
					return
				}

				log.WithError(err).Warn("Failed to validate JWT token")
				writeJSONError(w, http.StatusUnauthorized, "token_invalid", "Invalid authentication token")
				return
			}

			log.WithFields(logrus.Fields{
				"userId":            authCtx.UserID,
				"email":             authCtx.Email,
				"resourceOwnerId":   authCtx.ResourceOwnerID,
				"resourceOwnerName": authCtx.ResourceOwnerName,
				"platformAdmin":     authCtx.Roles.PlatformAdmin,
				"roleCount":         len(authCtx.Roles.AllRoles),
			}).Debug("JWT validated successfully")

			ctx = auth.SetAuthContext(ctx, authCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalJWTAuth validates tokens if present; passes through if absent
func OptionalJWTAuth(validator *auth.JWTValidator, log *logrus.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			authHeader := r.Header.Get("Authorization")
			userTokenHeader := r.Header.Get("X-User-Token")

			tokenString, err := auth.ExtractToken(authHeader, userTokenHeader)
			if err != nil {
				log.Debug("No JWT token provided, continuing without authentication")
				next.ServeHTTP(w, r)
				return
			}

			authCtx, err := validator.ExtractAuthContext(ctx, tokenString)
			if err != nil {
				log.WithError(err).Warn("Failed to validate JWT token, continuing without authentication")
				next.ServeHTTP(w, r)
				return
			}

			log.WithFields(logrus.Fields{
				"userId": authCtx.UserID,
				"email":  authCtx.Email,
			}).Debug("Optional JWT validated successfully")

			ctx = auth.SetAuthContext(ctx, authCtx)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeJSONError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	body, _ := json.Marshal(map[string]string{"error": code, "message": message})
	w.Write(body)
}

package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/clients/user"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/common/contextkeys"
	"github.com/tahion/ml_ab_test_service/services/api_gateway/internal/gateway"
)

type TokenVerifier interface {
	VerifyToken(ctx context.Context, token string) (*user.InternalUser, error)
}

func Auth(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				writeUnauthorized(w, "missing authorization header")
				return
			}

			u, err := verifier.VerifyToken(r.Context(), token)
			if err != nil {
				writeUnauthorized(w, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), contextkeys.UserIDKey, u.ID)
			ctx = context.WithValue(ctx, contextkeys.UserRoleKey, string(u.Role))
			ctx = context.WithValue(ctx, contextkeys.AuthTokenKey, token)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(allowed ...user.UserRole) func(http.Handler) http.Handler {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, r := range allowed {
		allowedSet[string(r)] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, ok := r.Context().Value(contextkeys.UserRoleKey).(string)
			if !ok || role == "" {
				writeForbidden(w, "no role in context")
				return
			}

			if _, allowed := allowedSet[role]; !allowed {
				writeForbidden(w, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func GetUserID(r *http.Request) (int, bool) {
	id, ok := r.Context().Value(contextkeys.UserIDKey).(int)
	return id, ok
}

func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

func writeUnauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", "Bearer")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"` + msg + `"}`))
}

func writeForbidden(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_, _ = w.Write([]byte(`{"error":"` + gateway.ErrForbidden.Error() + `","detail":"` + msg + `"}`))
}

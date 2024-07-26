package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/NisalSP9/To-Do-Web/user"
	"github.com/google/uuid"
)

const AuthUserID = "middleware.auth.userID"

func writeUnauthed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
}

func IsAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Whitelisted endpoints
		whitelistedPaths := []string{"/signup", "/login", "/health"}

		// Check if the request path is whitelisted
		for _, path := range whitelistedPaths {
			if r.URL.Path == path {
				next.ServeHTTP(w, r)
				return
			}
		}

		authorization := r.Header.Get("Authorization")

		// Check that the header begins with a prefix of Bearer
		if !strings.HasPrefix(authorization, "Bearer ") {
			writeUnauthed(w)
			return
		}

		// Pull out the token
		encodedToken := strings.TrimPrefix(authorization, "Bearer ")

		userHandler := &user.Handler{}

		userID, err := userHandler.Auth(encodedToken)

		if err != nil {
			writeUnauthed(w)
			return
		}

		if userID == uuid.Nil {
			writeUnauthed(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func GetAuthUserID(r *http.Request) (uuid.UUID, error) {

	authorization := r.Header.Get("Authorization")

	// Check that the header begins with a prefix of Bearer
	if !strings.HasPrefix(authorization, "Bearer ") {
		return uuid.Nil, errors.New(" prefix mismatch")
	}

	// Pull out the token
	encodedToken := strings.TrimPrefix(authorization, "Bearer ")

	userHandler := &user.Handler{}

	userID, err := userHandler.Auth(encodedToken)

	if err != nil {
		return uuid.Nil, err
	}

	if userID == uuid.Nil {
		return uuid.Nil, errors.New(" user not found")
	}

	return userID, nil
}

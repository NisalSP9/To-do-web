package middleware

import (
	"net/http"
)

// func EnsureAdmin(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		log.Println("Checking if user is admin")
// 		if !strings.Contains(r.Header.Get("Authorization"), "Admin") {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			w.Write([]byte(http.StatusText(http.StatusUnauthorized)))
// 			return
// 		}
// 		next.ServeHTTP(w, r)
// 	})
// }

// func LoadUser(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		log.Println("Loading user")
// 		next.ServeHTTP(w, r)
// 	})
// }

func AllowCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//log.Println("Enabling CORS")
		(w).Header().Set("Access-Control-Allow-Origin", "*")
		(w).Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		(w).Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Control-Allow-Headers, Authorization, X-Requested-With")
		next.ServeHTTP(w, r)
	})
}

// func CheckPermissions(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		log.Println("Checking Permissions")
// 		next.ServeHTTP(w, r)
// 	})
// }

package authcontroller

// func AuthorzationMiddleware(requiredScope int) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			tokenString := r.Header.Get("Authorization")
// 			if tokenString == "" {
// 				http.Error(w, "Missing token", http.StatusUnauthorized)
// 				return
// 			}
//
// 			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
// 				return secretKey, nil
// 			})
//
// 			if err != nil || !token.Valid {
// 				http.Error(w, "Invalid token", http.StatusUnauthorized)
// 				return
// 			}
//
// 			claims, ok := token.Claims.(jwt.MapClaims)
// 			if !ok {
// 				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
// 				return
// 			}
//
// 			scopeFloat, ok := claims["scope"].(float64)
// 			if !ok {
// 				http.Error(w, "Invalid scope in token claims", http.StatusUnauthorized)
// 				return
// 			}
//
// 			scope := int(scopeFloat)
//
// 			if !AuthorizeScope(scope, requiredScope) {
// 				http.Error(w, "Insufficient scope", http.StatusForbidden)
// 				return
// 			}
//
// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }

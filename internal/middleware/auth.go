package middleware

import (
	"net/http"
	"strings"

	"github.com/brayanzuritadev/StockManager/internal/services"
	"github.com/gin-gonic/gin"
)

// isPublicRoute returns true when the request is allowed without a JWT token.
// Public routes:
//   - POST /auth/login
//   - POST /auth/register
//   - GET  /products          (only when ?active=true query param is present)
//   - GET  /categories        (read-only catalog data)
//   - GET  /categories/{id}
//   - POST /carts             (anonymous customers create carts)
//   - POST /carts/{id}/items  (add items to an anonymous cart)
//   - POST /carts/{id}/share  (generate share link)
//   - GET  /carts/shared/{link} (store owner / anyone views shared cart)
func isPublicRoute(method, path string, query map[string][]string) bool {
	path = strings.Trim(path, "/")

	if method == http.MethodPost && (path == "auth/login" || path == "auth/register") {
		return true
	}

	if method == http.MethodGet {
		if path == "products" {
			vals, ok := query["active"]
			if ok && len(vals) > 0 && vals[0] == "true" {
				return true
			}
		}

		// All GET requests to /categories are public (read-only catalog data)
		if path == "categories" || strings.HasPrefix(path, "categories/") {
			return true
		}

		// GET /carts/shared/{link} — shared cart view (no auth needed)
		if strings.HasPrefix(path, "carts/shared/") {
			return true
		}
	}

	if method == http.MethodPost {
		// POST /carts — anonymous cart creation
		if path == "carts" {
			return true
		}
		// POST /carts/{id}/items  and  POST /carts/{id}/share
		parts := strings.SplitN(path, "/", 3)
		if len(parts) >= 2 && parts[0] == "carts" {
			if len(parts) == 3 && (parts[2] == "items" || parts[2] == "share") {
				return true
			}
		}
	}

	return false
}

// JWTAuth returns a Gin middleware that validates the Bearer token for all
// non-public routes. The jwtSecret is passed at setup time.
func JWTAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawPath := strings.TrimPrefix(c.Param("path"), "/")

		if isPublicRoute(c.Request.Method, rawPath, c.Request.URL.Query()) {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header must be 'Bearer <token>'"})
			return
		}

		_, err := services.ValidateJWT(parts[1], jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Next()
	}
}

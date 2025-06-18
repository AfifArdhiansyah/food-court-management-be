package middleware

import (
	"net/http"
	"strings"

	"foodcourt-backend/internal/models"
	"foodcourt-backend/pkg/auth"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	jwtService *auth.JWTService
}

func NewAuthMiddleware(jwtService *auth.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

func (a *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]
		claims, err := a.jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("kios_id", claims.KiosID)

		c.Next()
	}
}

func (a *AuthMiddleware) RequireRole(roles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User role not found in context",
			})
			c.Abort()
			return
		}

		role := userRole.(models.UserRole)
		for _, allowedRole := range roles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
		})
		c.Abort()
	}
}

func (a *AuthMiddleware) RequireKiosAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, _ := c.Get("role")
		role := userRole.(models.UserRole)

		// Cashier can access all kios
		if role == models.RoleCashier {
			c.Next()
			return
		}

		// Kios user can only access their own kios
		if role == models.RoleKios {
			userKiosID, exists := c.Get("kios_id")
			if !exists || userKiosID == nil {
				c.JSON(http.StatusForbidden, gin.H{
					"error": "Kios user must be assigned to a kios",
				})
				c.Abort()
				return
			}

			// Check if the requested kios matches user's kios
			requestedKiosID := c.Param("kios_id")
			if requestedKiosID != "" {
				userKiosIDUint := userKiosID.(*uint)
				if requestedKiosID != string(rune(*userKiosIDUint)) {
					c.JSON(http.StatusForbidden, gin.H{
						"error": "Access denied to this kios",
					})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

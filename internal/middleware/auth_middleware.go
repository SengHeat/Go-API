package middleware

import (
	"fmt"
	"go_land_api/internal/config"
	"go_land_api/internal/models"
	"go_land_api/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	jwt "github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware holds config and db pointer
type AuthMiddleware struct {
	cfg *config.Config
	db  *gorm.DB
}

func NewAuthMiddleware(cfg *config.Config, db *gorm.DB) *AuthMiddleware {
	return &AuthMiddleware{cfg: cfg, db: db}
}

func (a *AuthMiddleware) MiddlewareFunc() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.GetHeader("Authorization")
		if h == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing auth header"})
			return
		}
		parts := strings.SplitN(h, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid auth header"})
			return
		}
		tokenStr := parts[1]

		token, err := jwt.ParseWithClaims(tokenStr, &utils.TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(a.cfg.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		claims, ok := token.Claims.(*utils.TokenClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		// preload roles & permissions
		var user models.User
		if err := a.db.Preload("Roles.Permissions").First(&user, claims.UserID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "server error"})
			}
			return
		}

		c.Set("currentUser", user)
		c.Next()
	}
}

// RequirePermission returns middleware that checks current user's roles for a permission name
func RequirePermission(db *gorm.DB, permName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get("currentUser")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}
		user, ok := val.(models.User)
		if !ok {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "server error"})
			return
		}

		// iterate roles -> permissions to find the perm
		for _, r := range user.Roles {
			for _, p := range r.Permissions {
				if p.Name == permName {
					c.Next()
					return
				}
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("permission '%s' required", permName)})
	}
}

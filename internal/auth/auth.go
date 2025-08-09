package auth

import (
	"go_land_api/internal/config"
	"go_land_api/internal/models"
	"go_land_api/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func Login(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		var user models.User
		if err := db.Where("username = ?", req.Username).Preload("Roles.Permissions").First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		token, err := utils.GenerateJWT(cfg.JWTSecret, cfg.JWTHours, user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't create token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func Register(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		hash, err := utils.HashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't hash password"})
			return
		}

		user := models.User{
			Username:     req.Username,
			Email:        req.Email,
			PasswordHash: hash,
		}
		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't create user"})
			return
		}

		// by default assign "user" role if exists
		var role models.Role
		if err := db.Where("name = ?", "user").First(&role).Error; err == nil {
			db.Model(&user).Association("Roles").Append(&role)
		}

		c.JSON(http.StatusCreated, gin.H{"message": "user created"})
	}
}

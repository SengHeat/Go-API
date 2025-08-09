package controllers

import (
	"net/http"

	"go_land_api/internal/config"
	"go_land_api/internal/models"
	"go_land_api/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthController struct {
	DB  *gorm.DB
	Cfg *config.Config
}

func (ac *AuthController) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't hash password"})
		return
	}

	user := models.User{Username: req.Username, Email: req.Email, PasswordHash: hash}
	if err := ac.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't create user"})
		return
	}

	var role models.Role
	if err := ac.DB.Where("name = ?", "user").First(&role).Error; err == nil {
		ac.DB.Model(&user).Association("Roles").Append(&role)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user created"})
}

func (ac *AuthController) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	var user models.User
	if err := ac.DB.Where("username = ?", req.Username).Preload("Roles.Permissions").First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if !utils.CheckPasswordHash(req.Password, user.PasswordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := utils.GenerateJWT(ac.Cfg.JWTSecret, ac.Cfg.JWTHours, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

package controllers

import (
	"net/http"

	"go_land_api/internal/models"
	"go_land_api/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController struct {
	DB *gorm.DB
}

func (uc *UserController) Me(c *gin.Context) {
	val, _ := c.Get("currentUser")
	user := val.(models.User)
	user.PasswordHash = ""
	c.JSON(http.StatusOK, user)
}

func (uc *UserController) ListUsers(c *gin.Context) {
	var users []models.User
	uc.DB.Preload("Roles.Permissions").Find(&users)
	for i := range users {
		users[i].PasswordHash = ""
	}
	c.JSON(http.StatusOK, users)
}

func (uc *UserController) CreateUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash, _ := utils.HashPassword(req.Password)
	user := models.User{Username: req.Username, Email: req.Email, PasswordHash: hash}

	if err := uc.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't create user"})
		return
	}

	if req.Role != "" {
		var r models.Role
		if err := uc.DB.Where("name = ?", req.Role).First(&r).Error; err == nil {
			uc.DB.Model(&user).Association("Roles").Append(&r)
		}
	}

	user.PasswordHash = ""
	c.JSON(http.StatusCreated, user)
}

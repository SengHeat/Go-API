package handlers

import (
	"go_land_api/internal/config"
	"go_land_api/internal/models"
	"go_land_api/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// wrapper to expose Login/Register using the auth package
import authpkg "go_land_api/internal/auth"

func Register(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return authpkg.Register(db, cfg)
}

func Login(db *gorm.DB, cfg *config.Config) gin.HandlerFunc {
	return authpkg.Login(db, cfg)
}

func Me(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		val, _ := c.Get("currentUser")
		user := val.(models.User)
		// hide password hash
		user.PasswordHash = ""
		c.JSON(http.StatusOK, user)
	}
}

// List users (requires user:read)
func ListUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []models.User
		db.Preload("Roles.Permissions").Find(&users)
		for i := range users {
			users[i].PasswordHash = ""
		}
		c.JSON(http.StatusOK, users)
	}
}

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"` // optional role name
}

func CreateUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		hash, _ := utils.HashPassword(req.Password)
		user := models.User{
			Username:     req.Username,
			Email:        req.Email,
			PasswordHash: hash,
		}
		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "couldn't create user"})
			return
		}
		if req.Role != "" {
			var r models.Role
			if err := db.Where("name = ?", req.Role).First(&r).Error; err == nil {
				db.Model(&user).Association("Roles").Append(&r)
			}
		}
		user.PasswordHash = ""
		c.JSON(http.StatusCreated, user)
	}
}

// -- Simple posts endpoint examples (Just in-memory or minimal persisted model)
type Post struct {
	ID      uint   `json:"id" gorm:"primarykey"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func ListPosts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var posts []Post
		// For demo: you could persist posts to table; for brevity return static
		posts = append(posts, Post{ID: 1, Title: "Hello", Content: "First post"})
		c.JSON(http.StatusOK, posts)
	}
}

type CreatePostReq struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

func CreatePost(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreatePostReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		p := Post{Title: req.Title, Content: req.Content}
		c.JSON(http.StatusCreated, p)
	}
}

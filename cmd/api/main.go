package main

import (
	"log"
	"os"

	"go_land_api/internal/config"
	"go_land_api/internal/controllers"
	"go_land_api/internal/db"
	"go_land_api/internal/middleware"
	"go_land_api/internal/models"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	database := db.Connect(cfg)

	err := database.AutoMigrate(&models.User{}, &models.Role{}, &models.Permission{}, &models.UserRole{}, &models.RolePermission{})
	if err != nil {
		log.Fatalf("migrate failed: %v", err)
	}

	models.SeedDefaults(database)

	r := gin.Default()
	authMiddleware := middleware.NewAuthMiddleware(cfg, database)

	// Controllers
	authCtrl := controllers.AuthController{DB: database, Cfg: cfg}
	userCtrl := controllers.UserController{DB: database}
	postCtrl := controllers.PostController{}

	// Public
	r.POST("/register", authCtrl.Register)
	r.POST("/login", authCtrl.Login)

	// Protected
	api := r.Group("/api")
	api.Use(authMiddleware.MiddlewareFunc())
	{
		api.GET("/me", userCtrl.Me)
		api.GET("/users", middleware.RequirePermission(database, "user:read"), userCtrl.ListUsers)
		api.POST("/users", middleware.RequirePermission(database, "user:create"), userCtrl.CreateUser)

		api.GET("/posts", middleware.RequirePermission(database, "post:read"), postCtrl.ListPosts)
		api.POST("/posts", middleware.RequirePermission(database, "post:create"), postCtrl.CreatePost)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}

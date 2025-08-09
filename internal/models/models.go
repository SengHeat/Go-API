package models

import (
	"errors"
	"go_land_api/internal/utils"
	"log"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	Username     string    `gorm:"uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"not null" json:"-"`
	Roles        []Role    `gorm:"many2many:user_roles" json:"roles,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Role struct {
	ID          uint         `gorm:"primarykey" json:"id"`
	Name        string       `gorm:"uniqueIndex;not null" json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
}

type Permission struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserRole struct {
	UserID uint `gorm:"primaryKey"`
	RoleID uint `gorm:"primaryKey"`
}

type RolePermission struct {
	RoleID       uint `gorm:"primaryKey"`
	PermissionID uint `gorm:"primaryKey"`
}

// SeedDefaults inserts a couple of roles/permissions & an admin user for convenience.
func SeedDefaults(db *gorm.DB) {
	// create permissions
	perms := []Permission{
		{Name: "user:read", Description: "Read users"},
		{Name: "user:create", Description: "Create users"},
		{Name: "post:create", Description: "Create posts"},
		{Name: "post:read", Description: "Read posts"},
	}

	for _, p := range perms {
		var existing Permission
		if err := db.Where("name = ?", p.Name).First(&existing).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.Create(&p).Error; err != nil {
				log.Printf("perm create err: %v", err)
			}
		}
	}

	// roles
	adminRole := Role{Name: "admin", Description: "Administrator"}
	userRole := Role{Name: "user", Description: "Regular user"}

	if err := db.FirstOrCreate(&adminRole, Role{Name: "admin"}).Error; err != nil {
		log.Printf("role err: %v", err)
	}
	if err := db.FirstOrCreate(&userRole, Role{Name: "user"}).Error; err != nil {
		log.Printf("role err: %v", err)
	}

	// attach permissions to admin
	var allPerms []Permission
	db.Find(&allPerms)
	if len(allPerms) > 0 {
		db.Model(&adminRole).Association("Permissions").Replace(allPerms)
		// give user role only read permissions (example)
		var readPerms []Permission
		db.Where("name LIKE ?", "%:read").Find(&readPerms)
		db.Model(&userRole).Association("Permissions").Replace(readPerms)
	}

	// create admin user if not present
	var admin User
	if err := db.Where("username = ?", "admin").First(&admin).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		// use simple password "password" here for seed - change in production
		pass, _ := utils.HashPassword("password")
		admin = User{
			Username:     "admin",
			Email:        "admin@example.com",
			PasswordHash: pass,
		}
		if err := db.Create(&admin).Error; err == nil {
			// attach admin role
			db.Model(&admin).Association("Roles").Append(&adminRole)
		}
	}
}

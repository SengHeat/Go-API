package repository

import (
	"go_land_api/internal/models"

	"gorm.io/gorm"
)

type UserRepo struct {
	DB *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{DB: db} }

func (r *UserRepo) Create(u *models.User) error {
	return r.DB.Create(u).Error
}

func (r *UserRepo) FindByEmail(email string) (*models.User, error) {
	var u models.User
	if err := r.DB.Preload("Roles.Permissions").Where("email = ?", email).First(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) FindByID(id uint) (*models.User, error) {
	var u models.User
	if err := r.DB.Preload("Roles.Permissions").First(&u, id).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) AllUsers() ([]models.User, error) {
	var list []models.User
	if err := r.DB.Preload("Roles").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

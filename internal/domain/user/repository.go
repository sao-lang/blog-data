package user

import (
	"github.com/google/uuid"
	gorm "gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(user *User) error {
	user.ID = uuid.NewString()
	return r.db.Create(&user).Error
}

func (r *UserRepository) FindByUserName(userName string) (*User, error) {
	var user User
	if err := r.db.Where("user_name = ?", userName).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Delete(user *User) error {
	return r.db.Delete(&user).Error
}

func (r *UserRepository) Update(userId string, userInfo *User) error {
	var user User
	err := r.db.First(&user, userId).Error
	if err != nil {
		return err
	}
	return r.db.Model(&user).Updates(&userInfo).Error
}

// func (r *UserRepository) FindUsers() ([]*User, error) {
// 	if err := r.db.F
// }

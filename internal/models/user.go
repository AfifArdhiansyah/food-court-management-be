package models

import (
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	RoleCashier UserRole = "cashier"
	RoleKios    UserRole = "kios"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"`
	FullName  string         `json:"full_name" gorm:"not null"`
	Role      UserRole       `json:"role" gorm:"not null"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	KiosID    *uint          `json:"kios_id,omitempty" gorm:"index"`
	Kios      *Kios          `json:"kios,omitempty" gorm:"foreignKey:KiosID"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string   `json:"username" binding:"required,min=3,max=50"`
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password" binding:"required,min=6"`
	FullName string   `json:"full_name" binding:"required,min=2,max=100"`
	Role     UserRole `json:"role" binding:"required,oneof=cashier kios"`
	KiosID   *uint    `json:"kios_id,omitempty"`
}

type UserResponse struct {
	ID       uint     `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	FullName string   `json:"full_name"`
	Role     UserRole `json:"role"`
	IsActive bool     `json:"is_active"`
	KiosID   *uint    `json:"kios_id,omitempty"`
	Kios     *Kios    `json:"kios,omitempty"`
}

func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:       u.ID,
		Username: u.Username,
		Email:    u.Email,
		FullName: u.FullName,
		Role:     u.Role,
		IsActive: u.IsActive,
		KiosID:   u.KiosID,
		Kios:     u.Kios,
	}
}

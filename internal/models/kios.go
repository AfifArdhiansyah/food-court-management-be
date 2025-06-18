package models

import (
	"time"

	"gorm.io/gorm"
)

type Kios struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Location    string         `json:"location"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	Users       []User         `json:"users,omitempty" gorm:"foreignKey:KiosID"`
	Menus       []Menu         `json:"menus,omitempty" gorm:"foreignKey:KiosID"`
	Orders      []Order        `json:"orders,omitempty" gorm:"foreignKey:KiosID"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type CreateKiosRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=500"`
	Location    string `json:"location" binding:"max=200"`
}

type UpdateKiosRequest struct {
	Name        string `json:"name" binding:"omitempty,min=2,max=100"`
	Description string `json:"description" binding:"omitempty,max=500"`
	Location    string `json:"location" binding:"omitempty,max=200"`
	IsActive    *bool  `json:"is_active"`
}

type KiosResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	IsActive    bool      `json:"is_active"`
	MenuCount   int64     `json:"menu_count"`
	OrderCount  int64     `json:"order_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (k *Kios) ToResponse() *KiosResponse {
	return &KiosResponse{
		ID:          k.ID,
		Name:        k.Name,
		Description: k.Description,
		Location:    k.Location,
		IsActive:    k.IsActive,
		MenuCount:   int64(len(k.Menus)),
		OrderCount:  int64(len(k.Orders)),
		CreatedAt:   k.CreatedAt,
		UpdatedAt:   k.UpdatedAt,
	}
}

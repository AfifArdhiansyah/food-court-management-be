package models

import (
	"time"

	"gorm.io/gorm"
)

type MenuCategory string

const (
	CategoryFood    MenuCategory = "food"
	CategoryDrink   MenuCategory = "drink"
	CategorySnack   MenuCategory = "snack"
	CategoryDessert MenuCategory = "dessert"
)

type Menu struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	KiosID      uint           `json:"kios_id" gorm:"not null;index"`
	Kios        Kios           `json:"kios" gorm:"foreignKey:KiosID"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Price       float64        `json:"price" gorm:"not null"`
	Category    MenuCategory   `json:"category" gorm:"not null"`
	ImageURL    string         `json:"image_url"`
	IsAvailable bool           `json:"is_available" gorm:"default:true"`
	OrderItems  []OrderItem    `json:"order_items,omitempty" gorm:"foreignKey:MenuID"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type CreateMenuRequest struct {
	KiosID      uint         `json:"kios_id,omitempty"` // Optional in JSON, will be set from URL parameter
	Name        string       `json:"name" binding:"required,min=2,max=100"`
	Description string       `json:"description" binding:"max=500"`
	Price       float64      `json:"price" binding:"required,gt=0"`
	Category    MenuCategory `json:"category" binding:"required,oneof=food drink snack dessert"`
	ImageURL    string       `json:"image_url" binding:"omitempty,url"`
	IsAvailable *bool        `json:"is_available"`
}

type UpdateMenuRequest struct {
	Name        string       `json:"name" binding:"omitempty,min=2,max=100"`
	Description string       `json:"description" binding:"omitempty,max=500"`
	Price       *float64     `json:"price" binding:"omitempty,gt=0"`
	Category    MenuCategory `json:"category" binding:"omitempty,oneof=food drink snack dessert"`
	ImageURL    string       `json:"image_url" binding:"omitempty,url"`
	IsAvailable *bool        `json:"is_available"`
}

type MenuResponse struct {
	ID          uint         `json:"id"`
	KiosID      uint         `json:"kios_id"`
	KiosName    string       `json:"kios_name"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Price       float64      `json:"price"`
	Category    MenuCategory `json:"category"`
	ImageURL    string       `json:"image_url"`
	IsAvailable bool         `json:"is_available"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

func (m *Menu) ToResponse() *MenuResponse {
	return &MenuResponse{
		ID:          m.ID,
		KiosID:      m.KiosID,
		KiosName:    m.Kios.Name,
		Name:        m.Name,
		Description: m.Description,
		Price:       m.Price,
		Category:    m.Category,
		ImageURL:    m.ImageURL,
		IsAvailable: m.IsAvailable,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

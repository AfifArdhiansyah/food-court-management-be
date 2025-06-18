package models

import (
	"time"

	"gorm.io/gorm"
)

type OrderStatus string

const (
	StatusPending   OrderStatus = "pending"   // Pesanan dibuat, belum dibayar
	StatusPaid      OrderStatus = "paid"      // Sudah dibayar, masuk queue kios
	StatusPreparing OrderStatus = "preparing" // Sedang disiapkan
	StatusReady     OrderStatus = "ready"     // Siap diambil
	StatusCompleted OrderStatus = "completed" // Sudah diambil
	StatusCancelled OrderStatus = "cancelled" // Dibatalkan
)

type PaymentMethod string

const (
	PaymentCash    PaymentMethod = "cash"
	PaymentCard    PaymentMethod = "card"
	PaymentDigital PaymentMethod = "digital"
)

type Order struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	QueueNumber   string         `json:"queue_number" gorm:"uniqueIndex;not null"`
	KiosID        uint           `json:"kios_id" gorm:"not null;index"`
	Kios          Kios           `json:"kios" gorm:"foreignKey:KiosID"`
	CustomerName  string         `json:"customer_name"`
	Status        OrderStatus    `json:"status" gorm:"default:pending"`
	TotalAmount   float64        `json:"total_amount" gorm:"not null"`
	PaymentMethod *PaymentMethod `json:"payment_method"`
	PaidAt        *time.Time     `json:"paid_at"`
	PreparedAt    *time.Time     `json:"prepared_at"`
	ReadyAt       *time.Time     `json:"ready_at"`
	CompletedAt   *time.Time     `json:"completed_at"`
	Notes         string         `json:"notes"`
	OrderItems    []OrderItem    `json:"order_items" gorm:"foreignKey:OrderID"`
	CreatedBy     uint           `json:"created_by" gorm:"not null"`
	Creator       User           `json:"creator" gorm:"foreignKey:CreatedBy"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

type OrderItem struct {
	ID       uint    `json:"id" gorm:"primaryKey"`
	OrderID  uint    `json:"order_id" gorm:"not null;index"`
	Order    Order   `json:"order" gorm:"foreignKey:OrderID"`
	MenuID   uint    `json:"menu_id" gorm:"not null;index"`
	Menu     Menu    `json:"menu" gorm:"foreignKey:MenuID"`
	Quantity int     `json:"quantity" gorm:"not null"`
	Price    float64 `json:"price" gorm:"not null"` // Harga saat order dibuat
	Subtotal float64 `json:"subtotal" gorm:"not null"`
	Notes    string  `json:"notes"`
}

type CreateOrderRequest struct {
	KiosID       uint                     `json:"kios_id,omitempty"` // Optional in JSON, will be set from URL
	CustomerName string                   `json:"customer_name" binding:"max=100"`
	Notes        string                   `json:"notes" binding:"max=500"`
	Items        []CreateOrderItemRequest `json:"items" binding:"required,min=1"`
}

type CreateOrderItemRequest struct {
	MenuID   uint   `json:"menu_id" binding:"required"`
	Quantity int    `json:"quantity" binding:"required,min=1"`
	Notes    string `json:"notes" binding:"max=200"`
}

type UpdateOrderStatusRequest struct {
	Status        OrderStatus    `json:"status" binding:"required,oneof=pending paid preparing ready completed cancelled"`
	PaymentMethod *PaymentMethod `json:"payment_method" binding:"omitempty,oneof=cash card digital"`
}

type OrderResponse struct {
	ID            uint                `json:"id"`
	QueueNumber   string              `json:"queue_number"`
	KiosID        uint                `json:"kios_id"`
	KiosName      string              `json:"kios_name"`
	CustomerName  string              `json:"customer_name"`
	Status        OrderStatus         `json:"status"`
	TotalAmount   float64             `json:"total_amount"`
	PaymentMethod *PaymentMethod      `json:"payment_method"`
	PaidAt        *time.Time          `json:"paid_at"`
	PreparedAt    *time.Time          `json:"prepared_at"`
	ReadyAt       *time.Time          `json:"ready_at"`
	CompletedAt   *time.Time          `json:"completed_at"`
	Notes         string              `json:"notes"`
	OrderItems    []OrderItemResponse `json:"order_items"`
	CreatedBy     uint                `json:"created_by"`
	CreatorName   string              `json:"creator_name"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

type OrderItemResponse struct {
	ID       uint    `json:"id"`
	MenuID   uint    `json:"menu_id"`
	MenuName string  `json:"menu_name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Subtotal float64 `json:"subtotal"`
	Notes    string  `json:"notes"`
}

func (o *Order) ToResponse() *OrderResponse {
	items := make([]OrderItemResponse, len(o.OrderItems))
	for i, item := range o.OrderItems {
		items[i] = OrderItemResponse{
			ID:       item.ID,
			MenuID:   item.MenuID,
			MenuName: item.Menu.Name,
			Quantity: item.Quantity,
			Price:    item.Price,
			Subtotal: item.Subtotal,
			Notes:    item.Notes,
		}
	}

	return &OrderResponse{
		ID:            o.ID,
		QueueNumber:   o.QueueNumber,
		KiosID:        o.KiosID,
		KiosName:      o.Kios.Name,
		CustomerName:  o.CustomerName,
		Status:        o.Status,
		TotalAmount:   o.TotalAmount,
		PaymentMethod: o.PaymentMethod,
		PaidAt:        o.PaidAt,
		PreparedAt:    o.PreparedAt,
		ReadyAt:       o.ReadyAt,
		CompletedAt:   o.CompletedAt,
		Notes:         o.Notes,
		OrderItems:    items,
		CreatedBy:     o.CreatedBy,
		CreatorName:   o.Creator.FullName,
		CreatedAt:     o.CreatedAt,
		UpdatedAt:     o.UpdatedAt,
	}
}

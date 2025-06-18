package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"foodcourt-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OrderHandler struct {
	db *gorm.DB
}

func NewOrderHandler(db *gorm.DB) *OrderHandler {
	return &OrderHandler{db: db}
}

func (h *OrderHandler) generateQueueNumber(kiosID uint) (string, error) {
	// Get today's date
	today := time.Now().Format("20060102")

	// Count orders for this kios today
	var count int64
	h.db.Model(&models.Order{}).
		Where("kios_id = ? AND DATE(created_at) = ?", kiosID, time.Now().Format("2006-01-02")).
		Count(&count)

	// Generate queue number: KIOS{kiosID}-{date}-{sequence}
	return fmt.Sprintf("K%d-%s-%03d", kiosID, today, count+1), nil
}

func (h *OrderHandler) Create(c *gin.Context) {
	// Get kios ID from URL parameter
	kiosID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid kios ID",
		})
		return
	}

	var req models.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Override kios ID from URL parameter
	req.KiosID = uint(kiosID)

	userID, _ := c.Get("user_id")

	// Verify kios exists
	var kios models.Kios
	if err := h.db.First(&kios, req.KiosID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Kios not found",
		})
		return
	}

	// Generate queue number
	queueNumber, err := h.generateQueueNumber(req.KiosID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate queue number",
		})
		return
	}

	// Start transaction
	tx := h.db.Begin()

	// Create order
	order := models.Order{
		QueueNumber:  queueNumber,
		KiosID:       req.KiosID,
		CustomerName: req.CustomerName,
		Status:       models.StatusPending,
		Notes:        req.Notes,
		CreatedBy:    userID.(uint),
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create order",
		})
		return
	}

	// Create order items and calculate total
	var totalAmount float64
	for _, item := range req.Items {
		var menu models.Menu
		if err := tx.First(&menu, item.MenuID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Menu with ID %d not found", item.MenuID),
			})
			return
		}

		if !menu.IsAvailable {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("Menu '%s' is not available", menu.Name),
			})
			return
		}

		subtotal := menu.Price * float64(item.Quantity)
		totalAmount += subtotal

		orderItem := models.OrderItem{
			OrderID:  order.ID,
			MenuID:   item.MenuID,
			Quantity: item.Quantity,
			Price:    menu.Price,
			Subtotal: subtotal,
			Notes:    item.Notes,
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create order item",
			})
			return
		}
	}

	// Update order total
	order.TotalAmount = totalAmount
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update order total",
		})
		return
	}

	tx.Commit()

	// Load complete order data
	h.db.Preload("Kios").Preload("OrderItems.Menu").Preload("Creator").First(&order, order.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Order created successfully",
		"data":    order.ToResponse(),
	})
}

func (h *OrderHandler) GetByKios(c *gin.Context) {
	kiosID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid kios ID",
		})
		return
	}

	var orders []models.Order
	query := h.db.Preload("Kios").Preload("OrderItems.Menu").Preload("Creator").
		Where("kios_id = ?", uint(kiosID))

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by date if provided
	if date := c.Query("date"); date != "" {
		query = query.Where("DATE(created_at) = ?", date)
	}

	// Order by created_at desc
	query = query.Order("created_at DESC")

	if err := query.Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch orders",
		})
		return
	}

	responses := make([]*models.OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = order.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"data": responses,
	})
}

func (h *OrderHandler) GetAll(c *gin.Context) {
	var orders []models.Order
	query := h.db.Preload("Kios").Preload("OrderItems.Menu").Preload("Creator")

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by date if provided
	if date := c.Query("date"); date != "" {
		query = query.Where("DATE(created_at) = ?", date)
	}

	// Order by created_at desc
	query = query.Order("created_at DESC")

	if err := query.Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch orders",
		})
		return
	}

	responses := make([]*models.OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = order.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"data": responses,
	})
}

func (h *OrderHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order ID",
		})
		return
	}

	var order models.Order
	if err := h.db.Preload("Kios").Preload("OrderItems.Menu").Preload("Creator").
		First(&order, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Order not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch order",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": order.ToResponse(),
	})
}

func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order ID",
		})
		return
	}

	var req models.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	var order models.Order
	if err := h.db.First(&order, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Order not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch order",
		})
		return
	}

	// Update status and timestamps
	now := time.Now()
	order.Status = req.Status

	switch req.Status {
	case models.StatusPaid:
		order.PaidAt = &now
		order.PaymentMethod = req.PaymentMethod
	case models.StatusPreparing:
		order.PreparedAt = &now
	case models.StatusReady:
		order.ReadyAt = &now
	case models.StatusCompleted:
		order.CompletedAt = &now
	}

	if err := h.db.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update order status",
		})
		return
	}

	// Load complete order data
	h.db.Preload("Kios").Preload("OrderItems.Menu").Preload("Creator").First(&order, order.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Order status updated successfully",
		"data":    order.ToResponse(),
	})
}

func (h *OrderHandler) GetQueue(c *gin.Context) {
	kiosID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid kios ID",
		})
		return
	}

	// Get orders that are paid, preparing, or ready (active queue)
	var orders []models.Order
	if err := h.db.Preload("Kios").Preload("OrderItems.Menu").Preload("Creator").
		Where("kios_id = ? AND status IN ?", uint(kiosID), []models.OrderStatus{
			models.StatusPaid,
			models.StatusPreparing,
			models.StatusReady,
		}).
		Order("created_at ASC").
		Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch queue",
		})
		return
	}

	responses := make([]*models.OrderResponse, len(orders))
	for i, order := range orders {
		responses[i] = order.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"data": responses,
	})
}

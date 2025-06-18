package handlers

import (
	"net/http"
	"strconv"

	"foodcourt-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MenuHandler struct {
	db *gorm.DB
}

func NewMenuHandler(db *gorm.DB) *MenuHandler {
	return &MenuHandler{db: db}
}

func (h *MenuHandler) GetByKios(c *gin.Context) {
	kiosID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid kios ID",
		})
		return
	}

	var menus []models.Menu
	query := h.db.Preload("Kios").Where("kios_id = ?", uint(kiosID))

	// Filter by category if provided
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	// Filter by availability if provided
	if available := c.Query("available"); available != "" {
		if available == "true" {
			query = query.Where("is_available = ?", true)
		} else if available == "false" {
			query = query.Where("is_available = ?", false)
		}
	}

	if err := query.Find(&menus).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch menus",
		})
		return
	}

	responses := make([]*models.MenuResponse, len(menus))
	for i, menu := range menus {
		responses[i] = menu.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"data": responses,
	})
}

func (h *MenuHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid menu ID",
		})
		return
	}

	var menu models.Menu
	if err := h.db.Preload("Kios").First(&menu, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Menu not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch menu",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": menu.ToResponse(),
	})
}

func (h *MenuHandler) Create(c *gin.Context) {
	// Get kios ID from URL parameter
	kiosID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid kios ID",
		})
		return
	}

	var req models.CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	// Override kios ID from URL parameter
	req.KiosID = uint(kiosID)

	// Verify kios exists
	var kios models.Kios
	if err := h.db.First(&kios, req.KiosID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Kios not found",
		})
		return
	}

	isAvailable := true
	if req.IsAvailable != nil {
		isAvailable = *req.IsAvailable
	}

	menu := models.Menu{
		KiosID:      req.KiosID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Category:    req.Category,
		ImageURL:    req.ImageURL,
		IsAvailable: isAvailable,
	}

	if err := h.db.Create(&menu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create menu",
		})
		return
	}

	// Load kios data
	h.db.Preload("Kios").First(&menu, menu.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Menu created successfully",
		"data":    menu.ToResponse(),
	})
}

func (h *MenuHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid menu ID",
		})
		return
	}

	var req models.UpdateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	var menu models.Menu
	if err := h.db.Preload("Kios").First(&menu, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Menu not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch menu",
		})
		return
	}

	// Update fields
	if req.Name != "" {
		menu.Name = req.Name
	}
	if req.Description != "" {
		menu.Description = req.Description
	}
	if req.Price != nil {
		menu.Price = *req.Price
	}
	if req.Category != "" {
		menu.Category = req.Category
	}
	if req.ImageURL != "" {
		menu.ImageURL = req.ImageURL
	}
	if req.IsAvailable != nil {
		menu.IsAvailable = *req.IsAvailable
	}

	if err := h.db.Save(&menu).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update menu",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Menu updated successfully",
		"data":    menu.ToResponse(),
	})
}

func (h *MenuHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid menu ID",
		})
		return
	}

	if err := h.db.Delete(&models.Menu{}, uint(id)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete menu",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Menu deleted successfully",
	})
}

package handlers

import (
	"net/http"
	"strconv"

	"foodcourt-backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type KiosHandler struct {
	db *gorm.DB
}

func NewKiosHandler(db *gorm.DB) *KiosHandler {
	return &KiosHandler{db: db}
}

func (h *KiosHandler) GetAll(c *gin.Context) {
	var kios []models.Kios
	if err := h.db.Preload("Menus").Preload("Orders").Find(&kios).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch kios",
		})
		return
	}

	responses := make([]*models.KiosResponse, len(kios))
	for i, k := range kios {
		responses[i] = k.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"data": responses,
	})
}

func (h *KiosHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid kios ID",
		})
		return
	}

	var kios models.Kios
	if err := h.db.Preload("Menus").Preload("Orders").First(&kios, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Kios not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch kios",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": kios.ToResponse(),
	})
}

func (h *KiosHandler) Create(c *gin.Context) {
	var req models.CreateKiosRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	kios := models.Kios{
		Name:        req.Name,
		Description: req.Description,
		Location:    req.Location,
		IsActive:    true,
	}

	if err := h.db.Create(&kios).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create kios",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Kios created successfully",
		"data":    kios.ToResponse(),
	})
}

func (h *KiosHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid kios ID",
		})
		return
	}

	var req models.UpdateKiosRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}

	var kios models.Kios
	if err := h.db.First(&kios, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Kios not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch kios",
		})
		return
	}

	// Update fields
	if req.Name != "" {
		kios.Name = req.Name
	}
	if req.Description != "" {
		kios.Description = req.Description
	}
	if req.Location != "" {
		kios.Location = req.Location
	}
	if req.IsActive != nil {
		kios.IsActive = *req.IsActive
	}

	if err := h.db.Save(&kios).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update kios",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Kios updated successfully",
		"data":    kios.ToResponse(),
	})
}

func (h *KiosHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid kios ID",
		})
		return
	}

	if err := h.db.Delete(&models.Kios{}, uint(id)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete kios",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Kios deleted successfully",
	})
}

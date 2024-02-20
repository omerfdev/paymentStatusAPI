package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Balance  int64  `json:"balance"`
}

type CreatePaymentRequest struct {
	UserID    int64  `json:"user_id" binding:"required"`
	TotalCost int64  `json:"total_cost" binding:"required"`
}

type PaymentService struct {
	db *gorm.DB
}

func NewPaymentService(db *gorm.DB) *PaymentService {
	return &PaymentService{db: db}
}

func (s *PaymentService) CreatePayment(c *gin.Context) {
	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user User
	if err := s.db.Where("id = ?", req.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	if req.TotalCost > user.Balance {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "Account balance is negative"})
		return
	}

	user.Balance -= req.TotalCost
	s.db.Save(&user)

	c.JSON(http.StatusOK, gin.H{"remaining_balance": user.Balance})
}

func main() {
	db, err := gorm.Open("sqlite3", "./test.db")
	if err != nil {
		panic("failed to connect database")
	}
	defer db.Close()

	db.AutoMigrate(&User{})

	service := NewPaymentService(db)
	router := gin.Default()
	router.POST("/payments", service.CreatePayment)

	router.Run(":8080")
}
package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Product represents a product in the database
type Product struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"size:255"`
	Description string `gorm:"size:500"`
	Price       float64
	CreatedAt   time.Time
}

// setupDatabase initializes the database connection and migrates the schema
func setupDatabase() *gorm.DB {
	dsn := "user=postgres dbname=postgres password=123456 host=localhost sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}

	// Auto-migrate the Product schema
	db.AutoMigrate(&Product{})

	return db
}

// main function to set up the router and start the application
func main() {
	db := setupDatabase()

	r := gin.Default()

	r.GET("/products", func(c *gin.Context) {
		// Get the last ID from the query parameters
		lastIDStr := c.Query("last_id")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		var lastID uint
		if lastIDStr != "" {
			lastIDParsed, err := strconv.Atoi(lastIDStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid last_id"})
				return
			}
			lastID = uint(lastIDParsed)
		}

		var products []Product
		err := db.Where("id > ?", lastID).
			Order("id asc").
			Limit(limit).
			Find(&products).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, products)
	})

	r.Run(":8080")
}

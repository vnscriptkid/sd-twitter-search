package main

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Product struct {
	ID    int
	Name  string
	Price int
}

func main() {
	dsn := "user=postgres dbname=postgres password=123456 host=localhost sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		// logger
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}

	db.AutoMigrate(&Product{})

	r := gin.Default()
	r.GET("/products", func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
		offset := (page - 1) * limit

		var products []Product

		// SELECT * FROM products LIMIT :limit OFFSET :offset;
		err := db.Offset(offset).Limit(limit).Find(&products).Error
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, products)
	})

	r.Run()
}

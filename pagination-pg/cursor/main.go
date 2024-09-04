package main

import (
	"fmt"
	"net/http"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

var db *gorm.DB
var err error
var cursorStore = make(map[string]*gorm.DB)
var cursorStoreM sync.Mutex

// Initialize database connection
func initDB() {
	dsn := "user=postgres dbname=postgres password=123456 host=localhost sslmode=disable"
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database")
	}
}

// Handler for cursor pagination
func fetchUsers(c *gin.Context) {
	cursorID := c.Query("cursor_id")
	limit := 2 // Fetch 2 records at a time by default

	cursorStoreM.Lock()
	defer cursorStoreM.Unlock()

	var tx *gorm.DB
	var err error

	if cursorID == "" {
		// No cursor_id provided, start new pagination
		tx = db.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
			return
		}

		cursorID = fmt.Sprintf("cursor_%d", len(cursorStore)+1)
		cursorStore[cursorID] = tx

		// Declare CURSOR
		err = tx.Exec(fmt.Sprintf("DECLARE %s CURSOR FOR SELECT name FROM users ORDER BY id ASC", cursorID)).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to declare cursor"})
			tx.Rollback()
			delete(cursorStore, cursorID)
			return
		}
	} else {
		// Use existing CURSOR
		tx = cursorStore[cursorID]
		if tx == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cursor_id"})
			return
		}
	}

	// Fetch records
	rows, err := tx.Raw(fmt.Sprintf("FETCH FORWARD %d FROM %s", limit, cursorID)).Rows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		tx.Rollback()
		delete(cursorStore, cursorID)
		return
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan user"})
			tx.Rollback()
			delete(cursorStore, cursorID)
			return
		}
		users = append(users, name)
	}

	// Check if there are more records
	if len(users) < limit {
		// No more records, clean up
		tx.Commit()
		delete(cursorStore, cursorID)
		c.JSON(http.StatusOK, gin.H{"users": users, "message": "End of pagination"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users, "cursor_id": cursorID})
}

func main() {
	initDB()
	r := gin.Default()
	r.GET("/users", fetchUsers)
	r.Run()
}

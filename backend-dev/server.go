package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"
)

var db *sql.DB

func seedDatabase(db *sql.DB) {
	tableQuery := `
	CREATE TABLE IF NOT EXISTS account (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		ticker TEXT,
		cash_balance REAL
	);`

	_, err := db.Exec(tableQuery)
	if err != nil {
		log.Fatal("Failed to create table:", err)
	}

	// 2. Check if the table is empty
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM account").Scan(&count)
	if err != nil {
		log.Fatal("Failed to check table count:", err)
	}

	// 3. Only push dummy data if the table is brand new/empty
	if count == 0 {
		_, err = db.Exec("INSERT INTO account (ticker, cash_balance) VALUES ('CORAL', 50000.00)")
		if err != nil {
			log.Fatal("Failed to seed dummy data:", err)
		}
		fmt.Println("Successfully seeded database with dummy data!")
	}
}

func main() {
	var err error

	db, err = sql.Open("sqlite", "./coral-xchange.db")

	if err != nil {
		panic(err)
	}
	defer db.Close()

	seedDatabase(db)

	r := gin.Default()

	// Health endpoints
	r.GET("/", welcome)
	r.GET("/ping", ping)

	// Versioned API group
	api := r.Group("/api/v1")
	{
		api.GET("/portfolio", getPortfolio)
		api.GET("/account", getAccount)
		api.GET("/trades", getTrades)

		api.POST("/trade", placeTrade)

		api.GET("/prices", getPrices)
	}

	r.Run(":8080")

}

// --------------------
// Basic Handlers
// --------------------

func welcome(c *gin.Context) {
	c.String(http.StatusOK, "Welcome to the Coral Xchange API!")
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

// --------------------
// Trading Endpoints
// --------------------

// GET /api/v1/portfolio
func getPortfolio(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"holdings":   []gin.H{},
		"totalValue": 0,
	})
}

// GET /api/v1/account
func getAccount(c *gin.Context) {
	var balance float64
	var ticker string

	// QueryRow is used when you expect exactly one result
	err := db.QueryRow("SELECT ticker, cash_balance FROM account LIMIT 1").Scan(&ticker, &balance)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not retrieve account balance",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ticker":      ticker,
		"cashBalance": balance,
	})
}

// GET /api/v1/trades
func getTrades(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"trades": []gin.H{},
	})
}

// POST /api/v1/trade
func placeTrade(c *gin.Context) {
	var req struct {
		Symbol   string `json:"symbol"`
		Side     string `json:"side"` // BUY or SELL
		Quantity int    `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	// Stub response
	c.JSON(http.StatusOK, gin.H{
		"status":        "FILLED",
		"symbol":        req.Symbol,
		"side":          req.Side,
		"quantity":      req.Quantity,
		"price":         100.00,
		"remainingCash": 9000.00,
	})
}

// GET /api/v1/prices?symbols=AAPL,GOOG
func getPrices(c *gin.Context) {
	symbols := c.Query("symbols")

	c.JSON(http.StatusOK, gin.H{
		"requested": symbols,
		"prices": gin.H{
			"AAPL": 100.00,
			"GOOG": 200.00,
		},
	})
}

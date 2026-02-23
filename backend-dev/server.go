package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"
)

var db *sql.DB

func seedDatabase(db *sql.DB) {
	// 1. Create the account table if it doesn't exist
	tableQuery := `
	CREATE TABLE IF NOT EXISTS account (
		user_id INTEGER PRIMARY KEY AUTOINCREMENT,
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
		_, err = db.Exec("INSERT INTO account (user_id, cash_balance) VALUES (1, 100000.00)")
		if err != nil {
			log.Fatal("Failed to seed dummy account data:", err)
		}
		fmt.Println("Successfully seeded database with dummy data!")
	}

	// Create holdings table if it doesn't exist
	holdingsQuery := `
	CREATE TABLE IF NOT EXISTS holdings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		ticker TEXT,
		quantity INTEGER,
		price REAL
	);`

	_, err = db.Exec(holdingsQuery)
	if err != nil {
		log.Fatal("Failed to create holdings table:", err)
	}

	var holdingsCount int
	err = db.QueryRow("SELECT COUNT(*) FROM holdings").Scan(&holdingsCount)
	if err != nil {
		log.Fatal("Failed to check holdings table count:", err)
	}

	if holdingsCount == 0 {
		_, err = db.Exec("INSERT INTO holdings (user_id, ticker, quantity, price) VALUES (1, 'AAPL', 50, 150.00)")
		if err != nil {
			log.Fatal("Failed to seed dummy holdings data:", err)
		}
		fmt.Println("Successfully seeded holdings table with dummy data!")
	}
}

func main() {
	var err error

	// Use a relative DB path and log the absolute path so we know which file is opened.
	dbPath := "./coral-xchange.db"
	absPath, _ := filepath.Abs(dbPath)
	fmt.Println("Opening DB at:", absPath)
	db, err = sql.Open("sqlite", dbPath)

	if err != nil {
		panic(err)
	}
	defer db.Close()

	seedDatabase(db)

	r := gin.Default()

	// Health endpoints
	r.GET("/", welcome)

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

// --------------------
// Trading Endpoints
// --------------------

// GET /api/v1/portfolio
func getPortfolio(c *gin.Context) {
	rows, err := db.Query(`SELECT ticker, quantity, price FROM holdings WHERE user_id = 1;`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer rows.Close()

	var holdings []gin.H
	var totalValue float64

	for rows.Next() {
		var ticker string
		var quantity int
		var price float64

		err := rows.Scan(&ticker, &quantity, &price)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		totalValue += float64(quantity) * price
		holdings = append(holdings, gin.H{
			"ticker":   ticker,
			"quantity": quantity,
			"price":    price,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"holdings":   holdings,
		"totalValue": totalValue,
	})
}

// GET /api/v1/account
func getAccount(c *gin.Context) {
	var balance float64
	var userID int

	// QueryRow is used when you expect exactly one result
	err := db.QueryRow("SELECT user_id, cash_balance FROM account LIMIT 1").Scan(&userID, &balance)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not retrieve account balance",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userID":      userID,
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

	var balance float64
	err := db.QueryRow("SELECT cash_balance FROM account WHERE user_id = 1").Scan(&balance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Could not retrieve account balance",
		})
		return
	}

	switch req.Side {
	case "BUY":
		// we'll check if the user has enough cash balance
		// For simplicity, assume a fixed price of $100 per share
		tradeCost := float64(req.Quantity) * 100.00
		if balance < tradeCost {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Insufficient funds",
			})
			return
		}

		_, err = db.Exec("UPDATE account SET cash_balance = cash_balance - ? WHERE user_id = 1", tradeCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not update account balance",
			})
			return
		}

		// If a holding for this ticker already exists, increment quantity; otherwise insert new row
		var existingQty int
		err = db.QueryRow("SELECT quantity FROM holdings WHERE user_id = 1 AND ticker = ?", req.Symbol).Scan(&existingQty)
		if err == sql.ErrNoRows {
			res, err := db.Exec("INSERT INTO holdings (user_id, ticker, quantity, price) VALUES (1, ?, ?, 100.00)", req.Symbol, req.Quantity)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not insert holdings"})
				return
			}
			if ra, _ := res.RowsAffected(); ra == 0 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "No rows inserted into holdings"})
				return
			}
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not query holdings"})
			return
		} else {
			res, err := db.Exec("UPDATE holdings SET quantity = quantity + ? WHERE user_id = 1 AND ticker = ?", req.Quantity, req.Symbol)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update holdings"})
				return
			}
			if ra, _ := res.RowsAffected(); ra == 0 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "No rows updated in holdings"})
				return
			}
		}
		balance = balance - tradeCost

	case "SELL":
		// we check if the user has enough holdings to sell
		var quantity int
		err := db.QueryRow("SELECT quantity FROM holdings WHERE user_id = 1 AND ticker = ?", req.Symbol).Scan(&quantity)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not retrieve holdings",
			})
			return
		}
		if quantity < req.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Insufficient holdings to sell",
			})
			return
		}
		tradeProceeds := float64(req.Quantity) * 100.00
		_, err = db.Exec("UPDATE account SET cash_balance = cash_balance + ? WHERE user_id = 1", tradeProceeds)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not update account balance",
			})
			return
		}
		_, err = db.Exec("UPDATE holdings SET quantity = quantity - ? WHERE user_id = 1 AND ticker = ?", req.Quantity, req.Symbol)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Could not update holdings",
			})
			return
		}
		balance = balance + tradeProceeds

	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid trade side. Must be 'BUY' or 'SELL'",
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
		"remainingCash": balance,
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

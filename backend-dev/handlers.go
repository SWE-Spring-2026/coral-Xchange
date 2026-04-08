package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------------------
// Basic handlers
// ---------------------------------------------------------------------------

func welcome(c *gin.Context) {
	c.String(http.StatusOK, "Welcome to the Coral Xchange API!")
}

// ---------------------------------------------------------------------------
// Trading endpoints
// ---------------------------------------------------------------------------

// GET /api/v1/portfolio
func getPortfolio(c *gin.Context) {
	userID := getUserID(c)

	rows, err := db.Query(
		`SELECT ticker, quantity, price FROM holdings WHERE user_id = ?`, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var holdings []gin.H
	var totalValue float64

	for rows.Next() {
		var ticker string
		var quantity int
		var price float64

		if err := rows.Scan(&ticker, &quantity, &price); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	userID := getUserID(c)

	var balance float64
	err := db.QueryRow(
		"SELECT cash_balance FROM account WHERE user_id = ?", userID,
	).Scan(&balance)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve account balance"})
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
	userID := getUserID(c)

	var req struct {
		Symbol   string `json:"symbol"`
		Side     string `json:"side"` // "BUY" or "SELL"
		Quantity int    `json:"quantity"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity must be greater than zero"})
		return
	}

	var balance float64
	err := db.QueryRow(
		"SELECT cash_balance FROM account WHERE user_id = ?", userID,
	).Scan(&balance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve account balance"})
		return
	}

	switch req.Side {
	case "BUY":
		// we'll check if the user has enough cash balance
		// For simplicity, assume a fixed price of $100 per share
		tradeCost := float64(req.Quantity) * 100.00
		if balance < tradeCost {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds"})
			return
		}

		_, err = db.Exec(
			"UPDATE account SET cash_balance = cash_balance - ? WHERE user_id = ?",
			tradeCost, userID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update account balance"})
			return
		}

		var existingQty int
		err = db.QueryRow(
			"SELECT quantity FROM holdings WHERE user_id = ? AND ticker = ?", userID, req.Symbol,
		).Scan(&existingQty)

		if err == sql.ErrNoRows {
			res, err := db.Exec(
				"INSERT INTO holdings (user_id, ticker, quantity, price) VALUES (?, ?, ?, 100.00)",
				userID, req.Symbol, req.Quantity,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not insert holding"})
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
			res, err := db.Exec(
				"UPDATE holdings SET quantity = quantity + ? WHERE user_id = ? AND ticker = ?",
				req.Quantity, userID, req.Symbol,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update holding"})
				return
			}
			if ra, _ := res.RowsAffected(); ra == 0 {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "No rows updated in holdings"})
				return
			}
		}
		balance -= tradeCost

	case "SELL":
		var quantity int
		err := db.QueryRow(
			"SELECT quantity FROM holdings WHERE user_id = ? AND ticker = ?", userID, req.Symbol,
		).Scan(&quantity)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You do not own any shares of " + req.Symbol})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve holdings"})
			return
		}

		if quantity < req.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient holdings to sell"})
			return
		}

		tradeProceeds := float64(req.Quantity) * 100.00
		_, err = db.Exec(
			"UPDATE account SET cash_balance = cash_balance + ? WHERE user_id = ?",
			tradeProceeds, userID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update account balance"})
			return
		}
		_, err = db.Exec(
			"UPDATE holdings SET quantity = quantity - ? WHERE user_id = ? AND ticker = ?",
			req.Quantity, userID, req.Symbol,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update holdings"})
			return
		}
		balance += tradeProceeds

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid trade side. Must be 'BUY' or 'SELL'."})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        "FILLED",
		"symbol":        req.Symbol,
		"side":          req.Side,
		"quantity":      req.Quantity,
		"price":         100.00,
		"remainingCash": balance,
	})
}

// ---------------------------------------------------------------------------
// External market data endpoints
// ---------------------------------------------------------------------------

type SearchResponse struct {
	Data []struct {
		Symbol string `json:"symbol"`
		Name   string `json:"name"`
	} `json:"data"`
}

// GET /api/v1/searchStocks?query=<term>
func searchStocks(c *gin.Context) {
	apiToken := os.Getenv("STOCK_API_KEY")
	searchTerm := c.Query("query")

	apiURL := fmt.Sprintf(
		"https://api.stockdata.org/v1/entity/search?search=%s&api_token=%s",
		searchTerm, apiToken,
	)
	resp, err := http.Get(apiURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reach stock API"})
		return
	}
	defer resp.Body.Close()

	var searchRes SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchRes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse search results"})
		return
	}

	if len(searchRes.Data) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No stocks found"})
		return
	}

	c.JSON(http.StatusOK, searchRes.Data)
}

type QuoteResponse struct {
	Data []struct {
		Ticker        string  `json:"ticker"`
		Name          string  `json:"name"`
		Price         float64 `json:"price"`
		DayHigh       float64 `json:"day_high"`
		DayLow        float64 `json:"day_low"`
		DayOpen       float64 `json:"day_open"`
		DayChange     float64 `json:"day_change"`
		Volume        int64   `json:"volume"`
		YearHigh      float64 `json:"52_week_high"`
		YearLow       float64 `json:"52_week_low"`
		LastTradeTime string  `json:"last_trade_time"`
	} `json:"data"`
}

// GET /api/v1/quote/:ticker
func getStockQuote(c *gin.Context) {
	ticker := c.Param("ticker")
	apiToken := os.Getenv("STOCK_API_KEY")
	apiURL := fmt.Sprintf(
		"https://api.stockdata.org/v1/data/quote?symbols=%s&api_token=%s",
		ticker, apiToken,
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Stock API unreachable"})
		return
	}
	defer resp.Body.Close()

	var quoteRes QuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&quoteRes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse quote data"})
		return
	}

	if len(quoteRes.Data) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticker not found"})
		return
	}

	c.JSON(http.StatusOK, quoteRes.Data[0])
}

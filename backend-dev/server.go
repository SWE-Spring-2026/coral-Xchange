package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
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
	c.JSON(http.StatusOK, gin.H{
		"cashBalance": 10000.00,
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

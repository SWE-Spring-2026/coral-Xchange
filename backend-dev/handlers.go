package main

import (
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

	// Aggregate all lots per ticker, computing total quantity and average cost basis.
	rows, err := db.Query(
		`SELECT
			ticker,
			SUM(quantity)                              AS total_qty,
			SUM(quantity * price) / SUM(quantity)      AS avg_price
		FROM holdings
		WHERE user_id = ?
		GROUP BY ticker`,
		userID,
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
		var avgPrice float64

		if err := rows.Scan(&ticker, &quantity, &avgPrice); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		positionValue := float64(quantity) * avgPrice
		totalValue += positionValue

		holdings = append(holdings, gin.H{
			"ticker":        ticker,
			"quantity":      quantity,
			"avgCostBasis":  avgPrice,
			"positionValue": positionValue,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"holdings":   holdings,
		"totalValue": totalValue,
	})
}

// GET /api/v1/holdings
// Returns every individual purchase lot so the client can see per-lot cost basis.
func getHoldings(c *gin.Context) {
	userID := getUserID(c)

	rows, err := db.Query(
		`SELECT ticker, quantity, price, purchased_at
		FROM holdings
		WHERE user_id = ?
		ORDER BY ticker ASC, purchased_at ASC`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var lots []gin.H
	for rows.Next() {
		var ticker, purchasedAt string
		var quantity int
		var price float64

		if err := rows.Scan(&ticker, &quantity, &price, &purchasedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		lots = append(lots, gin.H{
			"ticker":      ticker,
			"quantity":    quantity,
			"price":       price,
			"purchasedAt": purchasedAt,
		})
	}

	// Return an empty array rather than null if there are no lots yet.
	if lots == nil {
		lots = []gin.H{}
	}

	c.JSON(http.StatusOK, gin.H{"lots": lots})
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
// Returns the full sell history with realized P&L for each trade.
func getTrades(c *gin.Context) {
	userID := getUserID(c)

	rows, err := db.Query(
		`SELECT ticker, quantity, sell_price, cost_basis, realized_pnl, sold_at
		FROM trades
		WHERE user_id = ?
		ORDER BY sold_at DESC`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var trades []gin.H
	for rows.Next() {
		var ticker, soldAt string
		var quantity int
		var sellPrice, costBasis, realizedPnL float64

		if err := rows.Scan(&ticker, &quantity, &sellPrice, &costBasis, &realizedPnL, &soldAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		trades = append(trades, gin.H{
			"ticker":      ticker,
			"quantity":    quantity,
			"sellPrice":   sellPrice,
			"costBasis":   costBasis,
			"realizedPnL": realizedPnL,
			"soldAt":      soldAt,
		})
	}

	// Return an empty array rather than null if there are no trades yet.
	if trades == nil {
		trades = []gin.H{}
	}

	c.JSON(http.StatusOK, gin.H{"trades": trades})
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

func getQuoteData(ticker string) (QuoteResponse, error) {
	apiToken := os.Getenv("STOCK_API_KEY")

	apiURL := fmt.Sprintf(
		"https://api.stockdata.org/v1/data/quote?symbols=%s&api_token=%s",
		ticker, apiToken,
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		return QuoteResponse{}, fmt.Errorf("stock API unreachable: %w", err)
	}
	defer resp.Body.Close()

	var quoteRes QuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&quoteRes); err != nil {
		return QuoteResponse{}, fmt.Errorf("failed to parse quote data: %w", err)
	}

	return quoteRes, nil
}

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

	// Fetch real-time price before doing anything else.
	quoteRes, err := getQuoteData(req.Symbol)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	if len(quoteRes.Data) == 0 {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Ticker not found"})
		return
	}

	stockPrice := quoteRes.Data[0].Price

	var balance float64
	err = db.QueryRow(
		"SELECT cash_balance FROM account WHERE user_id = ?", userID,
	).Scan(&balance)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve account balance"})
		return
	}

	switch req.Side {

	// -----------------------------------------------------------------------
	// BUY — always insert a new lot so we preserve the purchase price.
	// -----------------------------------------------------------------------
	case "BUY":
		tradeCost := float64(req.Quantity) * stockPrice
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

		// Each purchase becomes its own row so per-lot cost basis is preserved.
		_, err = db.Exec(
			"INSERT INTO holdings (user_id, ticker, quantity, price) VALUES (?, ?, ?, ?)",
			userID, req.Symbol, req.Quantity, stockPrice,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not insert holding"})
			return
		}

		balance -= tradeCost

	// -----------------------------------------------------------------------
	// SELL — consume lots oldest-first (FIFO) at today's market price.
	//
	// For each lot consumed we calculate:
	//   realized P&L = (sell price - lot buy price) × shares sold from that lot
	// The total across all lots consumed is recorded in the trades table.
	// -----------------------------------------------------------------------
	case "SELL":
		// Confirm the user owns enough shares across all lots.
		var totalQty int
		err := db.QueryRow(
			"SELECT COALESCE(SUM(quantity), 0) FROM holdings WHERE user_id = ? AND ticker = ?",
			userID, req.Symbol,
		).Scan(&totalQty)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve holdings"})
			return
		}
		if totalQty == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You do not own any shares of " + req.Symbol})
			return
		}
		if totalQty < req.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient holdings to sell"})
			return
		}

		// Fetch lots oldest-first for FIFO consumption.
		// We also select each lot's buy price to compute realized P&L.
		lotRows, err := db.Query(
			`SELECT id, quantity, price FROM holdings
			WHERE user_id = ? AND ticker = ?
			ORDER BY purchased_at ASC, id ASC`,
			userID, req.Symbol,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve holdings"})
			return
		}

		type lot struct {
			id       int
			qty      int
			buyPrice float64
		}
		var lots []lot
		for lotRows.Next() {
			var l lot
			if err := lotRows.Scan(&l.id, &l.qty, &l.buyPrice); err != nil {
				lotRows.Close()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not read holding"})
				return
			}
			lots = append(lots, l)
		}
		lotRows.Close()

		// Walk lots FIFO, consuming shares and accumulating cost basis + realized P&L.
		remaining := req.Quantity
		var totalCostBasis float64
		var totalRealizedPnL float64

		for _, l := range lots {
			if remaining == 0 {
				break
			}

			// How many shares are we taking from this lot?
			sharesFromLot := l.qty
			if sharesFromLot > remaining {
				sharesFromLot = remaining
			}

			// Tally this lot's contribution to the overall cost basis and P&L.
			totalCostBasis += float64(sharesFromLot) * l.buyPrice
			totalRealizedPnL += float64(sharesFromLot) * (stockPrice - l.buyPrice)

			if l.qty <= remaining {
				// Entire lot consumed — remove the row.
				if _, err := db.Exec("DELETE FROM holdings WHERE id = ?", l.id); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete holding lot"})
					return
				}
			} else {
				// Lot partially consumed — reduce its quantity.
				if _, err := db.Exec(
					"UPDATE holdings SET quantity = quantity - ? WHERE id = ?",
					remaining, l.id,
				); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update holding lot"})
					return
				}
			}

			remaining -= sharesFromLot
		}

		// Record the completed sell trade with its realized P&L.
		_, err = db.Exec(
			`INSERT INTO trades (user_id, ticker, quantity, sell_price, cost_basis, realized_pnl)
			VALUES (?, ?, ?, ?, ?, ?)`,
			userID, req.Symbol, req.Quantity, stockPrice, totalCostBasis, totalRealizedPnL,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not record trade"})
			return
		}

		// Credit the account at the current market price.
		tradeProceeds := float64(req.Quantity) * stockPrice
		_, err = db.Exec(
			"UPDATE account SET cash_balance = cash_balance + ? WHERE user_id = ?",
			tradeProceeds, userID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update account balance"})
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
		"price":         stockPrice,
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

// GET /api/v1/quote/:ticker
func getStockQuote(c *gin.Context) {
	ticker := c.Param("ticker")

	quoteRes, err := getQuoteData(ticker)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(quoteRes.Data) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticker not found"})
		return
	}

	c.JSON(http.StatusOK, quoteRes.Data[0])
}
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	cors "github.com/rs/cors/wrapper/gin"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB
var jwtSecret []byte

// ---------------------------------------------------------------------------
// JWT
// ---------------------------------------------------------------------------

// Claims defines the structure of the JSON Web Token payload for authenticated users.
// It is instantiated during login and parsed/validated on each protected request by authMiddleware.
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// authMiddleware validates the Bearer token on every protected route and
// injects userID / username into the Gin context for downstream handlers.
func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required (Bearer <token>)"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}

		// Parse and validate the token
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}

// helper that pulls the authenticated user's ID
// from the Gin context (set by authMiddleware).
func getUserID(c *gin.Context) int {
	id, _ := c.Get("userID")
	return id.(int)
}

// ---------------------------------------------------------------------------
// Database initialisation
// ---------------------------------------------------------------------------

func seedDatabase(db *sql.DB) {
}
func initDatabase(database *sql.DB) {
	// Enable foreign-key enforcement (off by default in SQLite).
	_, err := database.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		log.Fatal("Failed to enable foreign keys:", err)
	}

	queries := []string{
		// Users table — the anchor for every other table.
		`CREATE TABLE IF NOT EXISTS users (
			id            INTEGER PRIMARY KEY AUTOINCREMENT,
			username      TEXT UNIQUE NOT NULL,
			email         TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at    DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// One account row per user, created automatically on registration.
		`CREATE TABLE IF NOT EXISTS account (
			user_id      INTEGER PRIMARY KEY,
			cash_balance REAL NOT NULL DEFAULT 100000.00,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,

		// Each holding is scoped to a user.
		`CREATE TABLE IF NOT EXISTS holdings (
			id       INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id  INTEGER NOT NULL,
			ticker   TEXT    NOT NULL,
			quantity INTEGER NOT NULL,
			price    REAL    NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
	}

	for _, q := range queries {
		if _, err := database.Exec(q); err != nil {
			log.Fatal("Failed to run init query:", err)
		}
	}

	fmt.Println("Database initialised successfully.")
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file. Ensure STOCK_API_KEY and JWT_SECRET are set.")
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET must be set in .env")
	}
	jwtSecret = []byte(secret)

	// Use a relative DB path and log the absolute path so we know which file is opened.
	dbPath := "./coral-xchange.db"
	absPath, _ := filepath.Abs(dbPath)
	fmt.Println("Opening DB at:", absPath)
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	initDatabase(db)

	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/", welcome)

	api := r.Group("/api/v1")
	{
		// ── Public auth routes ────────────────────────────────────────────
		auth := api.Group("/auth")
		{
			auth.POST("/register", register)
			auth.POST("/login", login)
		}

		// ── Protected routes (JWT required) ───────────────────────────────
		protected := api.Group("/")
		protected.Use(authMiddleware())
		{
			protected.GET("/auth/me", getMe)

			protected.GET("/portfolio", getPortfolio)
			protected.GET("/account", getAccount)
			protected.GET("/trades", getTrades)
			protected.POST("/trade", placeTrade)

			protected.GET("/searchStocks", searchStocks)
			protected.GET("/quote/:ticker", getStockQuote)
		}
	}

	r.Run(":8080")
}

// ---------------------------------------------------------------------------
// Auth handlers
// ---------------------------------------------------------------------------

// POST /api/v1/auth/register
func register(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username, email, and password are all required"})
		return
	}

	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password must be at least 8 characters"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}

	res, err := db.Exec(
		"INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)",
		req.Username, req.Email, string(hash),
	)
	if err != nil {
		// A UNIQUE constraint violation means the username or email is taken.
		c.JSON(http.StatusConflict, gin.H{"error": "Username or email is already in use"})
		return
	}

	userID, _ := res.LastInsertId()

	// Provision a fresh $100,000 account for the new user.
	_, err = db.Exec("INSERT INTO account (user_id, cash_balance) VALUES (?, 100000.00)", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create account for user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Registration successful",
		"userID":   userID,
		"username": req.Username,
	})
}

// POST /api/v1/auth/login
func login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	var userID int
	var passwordHash string
	err := db.QueryRow(
		"SELECT id, password_hash FROM users WHERE username = ?", req.Username,
	).Scan(&userID, &passwordHash)

	if err == sql.ErrNoRows {
		// Return the same message whether the username or password is wrong
		// to avoid leaking which field is incorrect.
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve user"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// JWT generation, with expiry set to 24 hours from now. (this may have to be shortened)
	expiresAt := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:   userID,
		Username: req.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":     tokenString,
		"expiresAt": expiresAt,
		"userID":    userID,
		"username":  req.Username,
	})
}

// Consider renaming to profile or userInfo
// GET /api/v1/auth/me
func getMe(c *gin.Context) {
	userID := getUserID(c)

	var email, createdAt string
	var username string
	err := db.QueryRow(
		"SELECT username, email, created_at FROM users WHERE id = ?", userID,
	).Scan(&username, &email, &createdAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve user info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userID":    userID,
		"username":  username,
		"email":     email,
		"createdAt": createdAt,
	})
}

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

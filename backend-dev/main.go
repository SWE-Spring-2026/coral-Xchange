package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"
	"github.com/joho/godotenv"
	cors "github.com/rs/cors/wrapper/gin"
)

var db *sql.DB
var jwtSecret []byte

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

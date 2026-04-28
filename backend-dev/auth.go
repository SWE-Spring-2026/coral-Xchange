package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

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
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}
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
// Auth handlers
// ---------------------------------------------------------------------------

// POST /api/v1/auth/register
func register(c *gin.Context) {
	var req struct {
		Name     string `json:"name"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.Name == "" || req.Username == "" || req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name, username, email, and password are all required"})
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
		"INSERT INTO users (name, username, email, password_hash) VALUES (?, ?, ?, ?)",
		req.Name, req.Username, req.Email, string(hash),
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
		"name":     req.Name,
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

	var name, email, createdAt string
	var username string
	err := db.QueryRow(
		"SELECT name, username, email, created_at FROM users WHERE id = ?", userID,
	).Scan(&name, &username, &email, &createdAt)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve user info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"userID":    userID,
		"name":      name,
		"username":  username,
		"email":     email,
		"createdAt": createdAt,
	})
}

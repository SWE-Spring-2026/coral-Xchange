package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"
	"github.com/golang-jwt/jwt/v5"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	testDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory DB: %v", err)
	}
	initDatabase(testDB)
	return testDB
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	jwtSecret = []byte("test-secret")
	r := gin.New()

	api := r.Group("/api/v1")

	auth := api.Group("/auth")
	auth.POST("/register", register)
	auth.POST("/login", login)

	protected := api.Group("/")
	protected.Use(authMiddleware())
	protected.GET("/auth/me", getMe)
	protected.GET("/portfolio", getPortfolio)
	protected.GET("/account", getAccount)
	protected.GET("/trades", getTrades)
	protected.POST("/trade", placeTrade)
	protected.GET("/searchStocks", searchStocks)
	protected.GET("/quote/:ticker", getStockQuote)

	return r
}

// registerTestUser registers a user and returns their JWT. Every test that
// hits a protected endpoint should call this first.
func registerTestUser(t *testing.T, r *gin.Engine) string {
	t.Helper()
	doRequest(r, "POST", "/api/v1/auth/register", map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
		"password": "password123",
	})

	w := doRequest(r, "POST", "/api/v1/auth/login", map[string]interface{}{
		"username": "testuser",
		"password": "password123",
	})

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	token, ok := resp["token"].(string)
	if !ok {
		t.Fatal("registerTestUser: login did not return a token — check register/login handlers")
	}
	return token
}

// registerTestUser2 registers a second distinct user, used for isolation tests.
func registerTestUser2(t *testing.T, r *gin.Engine) string {
	t.Helper()
	doRequest(r, "POST", "/api/v1/auth/register", map[string]interface{}{
		"username": "otheruser",
		"email":    "other@example.com",
		"password": "password456",
	})

	w := doRequest(r, "POST", "/api/v1/auth/login", map[string]interface{}{
		"username": "otheruser",
		"password": "password456",
	})

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	token, ok := resp["token"].(string)
	if !ok {
		t.Fatal("registerTestUser2: login did not return a token")
	}
	return token
}

// doRequest sends an HTTP request to the router. Pass a JWT as the last
// argument for protected routes; omit it or pass "" for public routes.
func doRequest(r *gin.Engine, method, path string, body interface{}, token ...string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, _ := http.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if len(token) > 0 && token[0] != "" {
		req.Header.Set("Authorization", "Bearer "+token[0])
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ---------------------------------------------------------------------------
// initDatabase
// ---------------------------------------------------------------------------

func TestInitDatabase_CreatesUsersTable(t *testing.T) {
	testDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()
	initDatabase(testDB)

	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Errorf("users table does not exist after initDatabase: %v", err)
	}
}

func TestInitDatabase_CreatesAccountTable(t *testing.T) {
	testDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()
	initDatabase(testDB)

	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM account").Scan(&count)
	if err != nil {
		t.Errorf("account table does not exist after initDatabase: %v", err)
	}
}

func TestInitDatabase_CreatesHoldingsTable(t *testing.T) {
	testDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()
	initDatabase(testDB)

	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM holdings").Scan(&count)
	if err != nil {
		t.Errorf("holdings table does not exist after initDatabase: %v", err)
	}
}

func TestInitDatabase_StartsEmpty(t *testing.T) {
	testDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()
	initDatabase(testDB)

	var userCount, accountCount, holdingsCount int
	testDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	testDB.QueryRow("SELECT COUNT(*) FROM account").Scan(&accountCount)
	testDB.QueryRow("SELECT COUNT(*) FROM holdings").Scan(&holdingsCount)

	if userCount != 0 || accountCount != 0 || holdingsCount != 0 {
		t.Errorf("expected empty tables after init, got users=%d account=%d holdings=%d",
			userCount, accountCount, holdingsCount)
	}
}

func TestInitDatabase_Idempotent(t *testing.T) {
	testDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()

	initDatabase(testDB)
	initDatabase(testDB) // second call should not panic or error

	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Errorf("tables broken after double initDatabase: %v", err)
	}
}

// ---------------------------------------------------------------------------
// POST /api/v1/auth/register
// ---------------------------------------------------------------------------

func TestRegister_Success(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "POST", "/api/v1/auth/register", map[string]interface{}{
		"username": "alice",
		"email":    "alice@example.com",
		"password": "securepass",
	})

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestRegister_ProvisionsCashBalance(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	doRequest(setupRouter(), "POST", "/api/v1/auth/register", map[string]interface{}{
		"username": "alice",
		"email":    "alice@example.com",
		"password": "securepass",
	})

	var balance float64
	err := db.QueryRow("SELECT cash_balance FROM account WHERE user_id = 1").Scan(&balance)
	if err != nil {
		t.Fatalf("no account row created after registration: %v", err)
	}
	if balance != 100000.00 {
		t.Errorf("expected starting balance 100000.00, got %.2f", balance)
	}
}

func TestRegister_ReturnsUserIDAndUsername(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "POST", "/api/v1/auth/register", map[string]interface{}{
		"username": "alice",
		"email":    "alice@example.com",
		"password": "securepass",
	})

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["userID"] == nil {
		t.Error("expected userID in registration response")
	}
	if resp["username"] != "alice" {
		t.Errorf("expected username 'alice', got %v", resp["username"])
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	doRequest(r, "POST", "/api/v1/auth/register", map[string]interface{}{
		"username": "alice",
		"email":    "alice@example.com",
		"password": "securepass",
	})

	// Same username, different email.
	w := doRequest(r, "POST", "/api/v1/auth/register", map[string]interface{}{
		"username": "alice",
		"email":    "alice2@example.com",
		"password": "securepass",
	})

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409 for duplicate username, got %d", w.Code)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	doRequest(r, "POST", "/api/v1/auth/register", map[string]interface{}{
		"username": "alice",
		"email":    "shared@example.com",
		"password": "securepass",
	})

	w := doRequest(r, "POST", "/api/v1/auth/register", map[string]interface{}{
		"username": "bob",
		"email":    "shared@example.com",
		"password": "securepass",
	})

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409 for duplicate email, got %d", w.Code)
	}
}

func TestRegister_MissingUsername(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "POST", "/api/v1/auth/register", map[string]interface{}{
		"email":    "alice@example.com",
		"password": "securepass",
	})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing username, got %d", w.Code)
	}
}

func TestRegister_MissingEmail(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "POST", "/api/v1/auth/register", map[string]interface{}{
		"username": "alice",
		"password": "securepass",
	})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing email, got %d", w.Code)
	}
}

func TestRegister_PasswordTooShort(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "POST", "/api/v1/auth/register", map[string]interface{}{
		"username": "alice",
		"email":    "alice@example.com",
		"password": "short",
	})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for short password, got %d", w.Code)
	}
}

func TestRegister_PasswordIsHashed(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	doRequest(setupRouter(), "POST", "/api/v1/auth/register", map[string]interface{}{
		"username": "alice",
		"email":    "alice@example.com",
		"password": "securepass",
	})

	var hash string
	db.QueryRow("SELECT password_hash FROM users WHERE username = 'alice'").Scan(&hash)

	if hash == "securepass" {
		t.Error("password was stored in plaintext — must be hashed")
	}
	if hash == "" {
		t.Error("no password hash found in DB after registration")
	}
}

// ---------------------------------------------------------------------------
// POST /api/v1/auth/login
// ---------------------------------------------------------------------------

func TestLogin_Success(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	doRequest(r, "POST", "/api/v1/auth/register", map[string]interface{}{
		"username": "alice",
		"email":    "alice@example.com",
		"password": "securepass",
	})

	w := doRequest(r, "POST", "/api/v1/auth/login", map[string]interface{}{
		"username": "alice",
		"password": "securepass",
	})

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on login, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestLogin_ReturnsToken(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	doRequest(r, "POST", "/api/v1/auth/register", map[string]interface{}{
		"username": "alice",
		"email":    "alice@example.com",
		"password": "securepass",
	})

	w := doRequest(r, "POST", "/api/v1/auth/login", map[string]interface{}{
		"username": "alice",
		"password": "securepass",
	})

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["token"] == nil || resp["token"] == "" {
		t.Error("expected a token in login response")
	}
	if resp["expiresAt"] == nil {
		t.Error("expected expiresAt in login response")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	doRequest(r, "POST", "/api/v1/auth/register", map[string]interface{}{
		"username": "alice",
		"email":    "alice@example.com",
		"password": "securepass",
	})

	w := doRequest(r, "POST", "/api/v1/auth/login", map[string]interface{}{
		"username": "alice",
		"password": "wrongpassword",
	})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for wrong password, got %d", w.Code)
	}
}

func TestLogin_UnknownUser(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "POST", "/api/v1/auth/login", map[string]interface{}{
		"username": "nobody",
		"password": "doesntmatter",
	})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for unknown user, got %d", w.Code)
	}
}

// Both wrong-username and wrong-password should return the same error message
// to avoid leaking which field was incorrect (user enumeration).
func TestLogin_WrongPasswordAndWrongUserSameError(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	doRequest(r, "POST", "/api/v1/auth/register", map[string]interface{}{
		"username": "alice",
		"email":    "alice@example.com",
		"password": "securepass",
	})

	wBadPass := doRequest(r, "POST", "/api/v1/auth/login", map[string]interface{}{
		"username": "alice",
		"password": "wrongpassword",
	})
	wBadUser := doRequest(r, "POST", "/api/v1/auth/login", map[string]interface{}{
		"username": "nobody",
		"password": "securepass",
	})

	var resp1, resp2 map[string]interface{}
	json.Unmarshal(wBadPass.Body.Bytes(), &resp1)
	json.Unmarshal(wBadUser.Body.Bytes(), &resp2)

	if resp1["error"] != resp2["error"] {
		t.Errorf("wrong-password and wrong-username should return the same error message — got '%v' vs '%v'", resp1["error"], resp2["error"])
	}
}

func TestLogin_MissingFields(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "POST", "/api/v1/auth/login", map[string]interface{}{
		"username": "alice",
	})

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing password, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// Auth middleware
// ---------------------------------------------------------------------------

func TestAuthMiddleware_NoToken(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "GET", "/api/v1/account", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with no token, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "GET", "/api/v1/account", nil, "this.is.not.a.valid.jwt")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with invalid token, got %d", w.Code)
	}
}

func TestAuthMiddleware_TokenSignedWithWrongSecret(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	// Sign a token with a different secret than the server uses.
	claims := &Claims{
		UserID:   1,
		Username: "hacker",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	badToken, _ := token.SignedString([]byte("wrong-secret"))

	w := doRequest(setupRouter(), "GET", "/api/v1/account", nil, badToken)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for token signed with wrong secret, got %d", w.Code)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	jwtSecret = []byte("test-secret")
	claims := &Claims{
		UserID:   1,
		Username: "alice",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // expired 1 hour ago
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredToken, _ := token.SignedString(jwtSecret)

	w := doRequest(setupRouter(), "GET", "/api/v1/account", nil, expiredToken)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for expired token, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// GET /api/v1/auth/me
// ---------------------------------------------------------------------------

func TestGetMe_ReturnsUserInfo(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	w := doRequest(r, "GET", "/api/v1/auth/me", nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["username"] != "testuser" {
		t.Errorf("expected username 'testuser', got %v", resp["username"])
	}
	if resp["email"] != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %v", resp["email"])
	}
	if resp["userID"] == nil {
		t.Error("expected userID in /me response")
	}
}

func TestGetMe_RequiresAuth(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "GET", "/api/v1/auth/me", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with no token, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// GET /api/v1/account
// ---------------------------------------------------------------------------

func TestGetAccount_ReturnsOK(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	w := doRequest(r, "GET", "/api/v1/account", nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetAccount_ReturnsCorrectBalance(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	w := doRequest(r, "GET", "/api/v1/account", nil, token)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["cashBalance"] != 100000.00 {
		t.Errorf("expected cashBalance 100000.00, got %v", resp["cashBalance"])
	}
	if resp["userID"] == nil {
		t.Error("expected userID in response")
	}
}

func TestGetAccount_RequiresAuth(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "GET", "/api/v1/account", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with no token, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// GET /api/v1/portfolio
// ---------------------------------------------------------------------------

func TestGetPortfolio_ReturnsOK(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	w := doRequest(r, "GET", "/api/v1/portfolio", nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetPortfolio_EmptyOnFreshAccount(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	w := doRequest(r, "GET", "/api/v1/portfolio", nil, token)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["totalValue"] != 0.0 && resp["totalValue"] != nil {
		t.Errorf("expected totalValue 0 for fresh account, got %v", resp["totalValue"])
	}
}

func TestGetPortfolio_ReflectsBoughtShares(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "AAPL", "side": "BUY", "quantity": 10,
	}, token)

	w := doRequest(r, "GET", "/api/v1/portfolio", nil, token)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	holdings, ok := resp["holdings"].([]interface{})
	if !ok || len(holdings) == 0 {
		t.Fatal("expected holdings after buying shares")
	}
	first := holdings[0].(map[string]interface{})
	if first["ticker"] != "AAPL" {
		t.Errorf("expected ticker AAPL, got %v", first["ticker"])
	}
}

func TestGetPortfolio_TotalValueIsCalculated(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	symbol := "AAPL"
	quantity := 10

	quoteRes, err := getQuoteData(symbol)
	if err != nil {
		t.Fatalf("failed to fetch quote data: %v", err)
	}

	if len(quoteRes.Data) == 0 {
		t.Fatalf("no quote data returned for %s", symbol)
	}

	price := quoteRes.Data[0].Price
	expectedValue := price * float64(quantity)

	// BUY shares
	doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": symbol,
		"side": "BUY",
		"quantity": quantity,
	}, token)

	w := doRequest(r, "GET", "/api/v1/portfolio", nil, token)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	totalValue := resp["totalValue"].(float64)

	if totalValue != expectedValue {
		t.Errorf(
			"expected totalValue %.2f, got %.2f (price %.2f)",
			expectedValue, totalValue, price,
		)
	}
}

func TestGetPortfolio_RequiresAuth(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "GET", "/api/v1/portfolio", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with no token, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// GET /api/v1/trades
// ---------------------------------------------------------------------------

func TestGetTrades_ReturnsOK(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	w := doRequest(r, "GET", "/api/v1/trades", nil, token)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetTrades_ReturnsEmptyList(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	w := doRequest(r, "GET", "/api/v1/trades", nil, token)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	trades, ok := resp["trades"].([]interface{})
	if !ok {
		t.Fatal("expected 'trades' array in response")
	}
	if len(trades) != 0 {
		t.Errorf("expected empty trades list, got %d items", len(trades))
	}
}

func TestGetTrades_RequiresAuth(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "GET", "/api/v1/trades", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with no token, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// POST /api/v1/trade — BUY
// ---------------------------------------------------------------------------

func TestPlaceTrade_BuySuccess(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	w := doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "TSLA", "side": "BUY", "quantity": 5,
	}, token)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestPlaceTrade_BuyDeductsBalance(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	symbol := "TSLA"
	quantity := 10

	// 🔥 Get REAL price (same logic as production code uses)
	quoteRes, err := getQuoteData(symbol)
	if err != nil {
		t.Fatalf("failed to fetch quote data: %v", err)
	}

	if len(quoteRes.Data) == 0 {
		t.Fatalf("no quote data returned for %s", symbol)
	}

	price := quoteRes.Data[0].Price
	expectedCost := price * float64(quantity)

	doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol":   symbol,
		"side":     "BUY",
		"quantity": quantity,
	}, token)

	var balance float64
	err = db.QueryRow(
		"SELECT cash_balance FROM account WHERE user_id = 1",
	).Scan(&balance)

	if err != nil {
		t.Fatalf("failed to fetch balance: %v", err)
	}

	expected := 100000.00 - expectedCost

	if balance != expected {
		t.Errorf(
			"expected balance %.2f after buy, got %.2f (price %.2f)",
			expected, balance, price,
		)
	}
}

func TestPlaceTrade_BuyAddsNewHolding(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "MSFT", "side": "BUY", "quantity": 3,
	}, token)

	var qty int
	err := db.QueryRow("SELECT quantity FROM holdings WHERE user_id = 1 AND ticker = 'MSFT'").Scan(&qty)
	if err != nil {
		t.Fatalf("MSFT holding not found after buy: %v", err)
	}
	if qty != 3 {
		t.Errorf("expected MSFT quantity 3, got %d", qty)
	}
}

func TestPlaceTrade_BuyAccumulatesExistingHolding(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	// Buy AAPL twice and verify quantities accumulate.
	doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "AAPL", "side": "BUY", "quantity": 10,
	}, token)
	doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "AAPL", "side": "BUY", "quantity": 5,
	}, token)

	var qty int
	db.QueryRow("SELECT quantity FROM holdings WHERE user_id = 1 AND ticker = 'AAPL'").Scan(&qty)

	if qty != 15 {
		t.Errorf("expected AAPL quantity 15 after two buys, got %d", qty)
	}
}

func TestPlaceTrade_BuyInsufficientFunds(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	// 100000 / 100 = 1000 max shares; request 1001
	w := doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "TSLA", "side": "BUY", "quantity": 1001,
	}, token)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for insufficient funds, got %d", w.Code)
	}
}

func TestPlaceTrade_BuyResponseFields(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	w := doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "NVDA", "side": "BUY", "quantity": 2,
	}, token)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	for _, field := range []string{"status", "symbol", "side", "quantity", "price", "remainingCash"} {
		if resp[field] == nil {
			t.Errorf("expected field '%s' in trade response", field)
		}
	}
	if resp["status"] != "FILLED" {
		t.Errorf("expected status FILLED, got %v", resp["status"])
	}
}

func TestPlaceTrade_RequiresAuth(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "AAPL", "side": "BUY", "quantity": 1,
	})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with no token, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// POST /api/v1/trade — SELL
// ---------------------------------------------------------------------------

func TestPlaceTrade_SellSuccess(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "AAPL", "side": "BUY", "quantity": 20,
	}, token)

	w := doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "AAPL", "side": "SELL", "quantity": 10,
	}, token)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestPlaceTrade_SellCreditsBalance(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	symbol := "AAPL"
	buyQty := 20
	sellQty := 10

	quoteRes, err := getQuoteData(symbol)
	if err != nil {
		t.Fatalf("failed to fetch quote data: %v", err)
	}

	if len(quoteRes.Data) == 0 {
		t.Fatalf("no quote data returned for %s", symbol)
	}

	price := quoteRes.Data[0].Price

	// BUY 20 shares
	doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol":   symbol,
		"side":     "BUY",
		"quantity": buyQty,
	}, token)

	// SELL 10 shares
	doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol":   symbol,
		"side":     "SELL",
		"quantity": sellQty,
	}, token)

	var balance float64
	err = db.QueryRow(
		"SELECT cash_balance FROM account WHERE user_id = 1",
	).Scan(&balance)

	if err != nil {
		t.Fatalf("failed to fetch balance: %v", err)
	}

	expected := 100000.00 - (float64(buyQty) * price) + (float64(sellQty) * price)

	if balance != expected {
		t.Errorf(
			"expected balance %.2f after buy+sell, got %.2f (price %.2f)",
			expected, balance, price,
		)
	}
}

func TestPlaceTrade_SellDecreasesHoldings(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "AAPL", "side": "BUY", "quantity": 30,
	}, token)
	doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "AAPL", "side": "SELL", "quantity": 10,
	}, token)

	var qty int
	db.QueryRow("SELECT quantity FROM holdings WHERE user_id = 1 AND ticker = 'AAPL'").Scan(&qty)

	if qty != 20 {
		t.Errorf("expected AAPL quantity 20 after buy 30 / sell 10, got %d", qty)
	}
}

func TestPlaceTrade_SellInsufficientHoldings(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "AAPL", "side": "BUY", "quantity": 5,
	}, token)

	w := doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "AAPL", "side": "SELL", "quantity": 100,
	}, token)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for insufficient holdings, got %d", w.Code)
	}
}

func TestPlaceTrade_SellTickerNotOwned(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	w := doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "GOOG", "side": "SELL", "quantity": 1,
	}, token)

	if w.Code == http.StatusOK {
		t.Error("expected non-200 when selling a ticker the user does not own")
	}
}

// ---------------------------------------------------------------------------
// POST /api/v1/trade — invalid inputs
// ---------------------------------------------------------------------------

func TestPlaceTrade_InvalidSide(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	w := doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "AAPL", "side": "HOLD", "quantity": 1,
	}, token)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid side, got %d", w.Code)
	}
}

func TestPlaceTrade_MalformedBody(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	req, _ := http.NewRequest("POST", "/api/v1/trade", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for malformed body, got %d", w.Code)
	}
}

func TestPlaceTrade_ZeroQuantity(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	w := doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "TSLA", "side": "BUY", "quantity": 0,
	}, token)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for zero quantity, got %d", w.Code)
	}
}

func TestPlaceTrade_NegativeQuantity(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	w := doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "TSLA", "side": "BUY", "quantity": -5,
	}, token)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for negative quantity, got %d", w.Code)
	}
}

// ---------------------------------------------------------------------------
// User data isolation
// ---------------------------------------------------------------------------

func TestUserIsolation_AccountsAreIndependent(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token1 := registerTestUser(t, r)
	token2 := registerTestUser2(t, r)

	// User 1 spends some money.
	doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "AAPL", "side": "BUY", "quantity": 100,
	}, token1)

	// User 2's balance should be untouched.
	w := doRequest(r, "GET", "/api/v1/account", nil, token2)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["cashBalance"] != 100000.00 {
		t.Errorf("user2 balance should be unaffected by user1's trade, got %v", resp["cashBalance"])
	}
}

func TestUserIsolation_PortfoliosAreIndependent(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token1 := registerTestUser(t, r)
	token2 := registerTestUser2(t, r)

	// User 1 buys AAPL.
	doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "AAPL", "side": "BUY", "quantity": 10,
	}, token1)

	// User 2 should see an empty portfolio.
	w := doRequest(r, "GET", "/api/v1/portfolio", nil, token2)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["totalValue"] != 0.0 && resp["totalValue"] != nil {
		t.Errorf("user2 portfolio should be empty, got totalValue=%v", resp["totalValue"])
	}
}

func TestUserIsolation_CannotSellAnotherUsersShares(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token1 := registerTestUser(t, r)
	token2 := registerTestUser2(t, r)

	// User 1 buys AAPL.
	doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "AAPL", "side": "BUY", "quantity": 10,
	}, token1)

	// User 2 tries to sell AAPL they don't own.
	w := doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": "AAPL", "side": "SELL", "quantity": 5,
	}, token2)

	if w.Code == http.StatusOK {
		t.Error("user2 should not be able to sell shares owned by user1")
	}
}

// ---------------------------------------------------------------------------
// External API handlers
// ---------------------------------------------------------------------------

func TestSearchStocks_MissingAPIKey(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	t.Setenv("STOCK_API_KEY", "invalid_test_key")

	w := doRequest(r, "GET", "/api/v1/searchStocks?query=AAPL", nil, token)
	if w.Code < 100 || w.Code > 599 {
		t.Errorf("unexpected status code %d from searchStocks", w.Code)
	}
}

func TestGetStockQuote_MissingAPIKey(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	t.Setenv("STOCK_API_KEY", "invalid_test_key")

	w := doRequest(r, "GET", "/api/v1/quote/AAPL", nil, token)
	if w.Code < 100 || w.Code > 599 {
		t.Errorf("unexpected status code %d from getStockQuote", w.Code)
	}
}

func TestSearchStocks_ParsesValidResponse(t *testing.T) {
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"data":[{"symbol":"AAPL","name":"Apple Inc"}]}`)
	}))
	defer mock.Close()

	raw := `{"data":[{"symbol":"AAPL","name":"Apple Inc"}]}`
	var sr SearchResponse
	if err := json.Unmarshal([]byte(raw), &sr); err != nil {
		t.Fatalf("SearchResponse failed to parse: %v", err)
	}
	if len(sr.Data) != 1 || sr.Data[0].Symbol != "AAPL" {
		t.Errorf("unexpected parsed data: %+v", sr.Data)
	}
}

func TestGetStockQuote_ParsesValidResponse(t *testing.T) {
	raw := `{"data":[{
		"ticker":"AAPL","name":"Apple Inc","price":174.50,
		"day_high":175.00,"day_low":173.00,"day_open":173.50,
		"day_change":1.00,"volume":1000000,
		"52_week_high":198.00,"52_week_low":124.00,
		"last_trade_time":"2024-01-01T16:00:00"
	}]}`

	var qr QuoteResponse
	if err := json.Unmarshal([]byte(raw), &qr); err != nil {
		t.Fatalf("QuoteResponse failed to parse: %v", err)
	}
	if qr.Data[0].Price != 174.50 {
		t.Errorf("expected price 174.50, got %v", qr.Data[0].Price)
	}
	if qr.Data[0].YearHigh != 198.00 {
		t.Errorf("expected 52-week high 198.00, got %v", qr.Data[0].YearHigh)
	}
}

// ---------------------------------------------------------------------------
// Multi-step balance consistency
// ---------------------------------------------------------------------------

func TestTradeSequence_BalanceConsistency(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()
	token := registerTestUser(t, r)

	symbol := "TSLA"

	quoteRes, err := getQuoteData(symbol)
	if err != nil {
		t.Fatalf("failed to fetch quote data: %v", err)
	}

	if len(quoteRes.Data) == 0 {
		t.Fatalf("no quote data returned for %s", symbol)
	}

	price := quoteRes.Data[0].Price

	// Step 1: BUY 50
	doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": symbol, "side": "BUY", "quantity": 50,
	}, token)

	// Step 2: SELL 20
	doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": symbol, "side": "SELL", "quantity": 20,
	}, token)

	// Step 3: BUY 10
	doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{
		"symbol": symbol, "side": "BUY", "quantity": 10,
	}, token)

	var balance float64
	err = db.QueryRow(
		"SELECT cash_balance FROM account WHERE user_id = 1",
	).Scan(&balance)

	if err != nil {
		t.Fatalf("failed to fetch balance: %v", err)
	}

	expected := 100000.00 -
		(50 * price) +
		(20 * price) -
		(10 * price)

	if balance != expected {
		t.Errorf(
			"expected balance %.2f, got %.2f (price %.2f)",
			expected, balance, price,
		)
	}

	var qty int
	err = db.QueryRow(
		"SELECT quantity FROM holdings WHERE user_id = 1 AND ticker = ?",
		symbol,
	).Scan(&qty)

	if err != nil {
		t.Fatalf("failed to fetch holdings: %v", err)
	}

	if qty != 40 {
		t.Errorf("expected TSLA quantity 40, got %d", qty)
	}
}
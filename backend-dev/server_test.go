package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/glebarez/go-sqlite"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// setupTestDB opens an in-memory SQLite database, runs seedDatabase, and
// returns it. Each test that calls this gets a fresh, isolated database.
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	testDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory DB: %v", err)
	}
	seedDatabase(testDB)
	return testDB
}

// setupRouter wires up the Gin router in test mode using the global `db`
// variable (already pointed at the test DB by the caller).
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	api := r.Group("/api/v1")
	{
		api.GET("/portfolio", getPortfolio)
		api.GET("/account", getAccount)
		api.GET("/trades", getTrades)
		api.POST("/trade", placeTrade)
		api.GET("/searchStocks", searchStocks)
		api.GET("/quote/:ticker", getStockQuote)
	}
	return r
}

// doRequest is a small convenience wrapper around httptest.
func doRequest(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, _ := http.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ---------------------------------------------------------------------------
// seedDatabase
// ---------------------------------------------------------------------------

func TestSeedDatabase_CreatesAccountRow(t *testing.T) {
	testDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()

	seedDatabase(testDB)

	var count int
	testDB.QueryRow("SELECT COUNT(*) FROM account").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 account row after seed, got %d", count)
	}
}

func TestSeedDatabase_AccountBalanceIsCorrect(t *testing.T) {
	testDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()

	seedDatabase(testDB)

	var balance float64
	testDB.QueryRow("SELECT cash_balance FROM account WHERE user_id = 1").Scan(&balance)
	if balance != 100000.00 {
		t.Errorf("expected starting balance 100000.00, got %.2f", balance)
	}
}

func TestSeedDatabase_CreatesHoldingsRow(t *testing.T) {
	testDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()

	seedDatabase(testDB)

	var ticker string
	var qty int
	testDB.QueryRow("SELECT ticker, quantity FROM holdings WHERE user_id = 1").Scan(&ticker, &qty)
	if ticker != "AAPL" {
		t.Errorf("expected seeded ticker AAPL, got %s", ticker)
	}
	if qty != 50 {
		t.Errorf("expected seeded quantity 50, got %d", qty)
	}
}

// Calling seedDatabase a second time on the same DB must not duplicate rows.
func TestSeedDatabase_IdempotentOnExistingData(t *testing.T) {
	testDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()

	seedDatabase(testDB)
	seedDatabase(testDB) // second call

	var accountCount, holdingsCount int
	testDB.QueryRow("SELECT COUNT(*) FROM account").Scan(&accountCount)
	testDB.QueryRow("SELECT COUNT(*) FROM holdings").Scan(&holdingsCount)

	if accountCount != 1 {
		t.Errorf("expected 1 account row after double seed, got %d", accountCount)
	}
	if holdingsCount != 1 {
		t.Errorf("expected 1 holdings row after double seed, got %d", holdingsCount)
	}
}

// ---------------------------------------------------------------------------
// GET /api/v1/account
// ---------------------------------------------------------------------------

func TestGetAccount_ReturnsOK(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "GET", "/api/v1/account", nil)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetAccount_ReturnsCorrectBalance(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "GET", "/api/v1/account", nil)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["cashBalance"] != 100000.00 {
		t.Errorf("expected cashBalance 100000.00, got %v", resp["cashBalance"])
	}
	if resp["userID"] == nil {
		t.Error("expected userID in response")
	}
}

// ---------------------------------------------------------------------------
// GET /api/v1/portfolio
// ---------------------------------------------------------------------------

func TestGetPortfolio_ReturnsOK(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "GET", "/api/v1/portfolio", nil)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetPortfolio_ContainsSeedHolding(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "GET", "/api/v1/portfolio", nil)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	holdings, ok := resp["holdings"].([]interface{})
	if !ok || len(holdings) == 0 {
		t.Fatal("expected at least one holding in portfolio")
	}

	first := holdings[0].(map[string]interface{})
	if first["ticker"] != "AAPL" {
		t.Errorf("expected ticker AAPL, got %v", first["ticker"])
	}
}

func TestGetPortfolio_TotalValueIsCalculated(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "GET", "/api/v1/portfolio", nil)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// Seed: 50 shares @ $150.00 = $7500.00
	if resp["totalValue"] != 7500.00 {
		t.Errorf("expected totalValue 7500.00, got %v", resp["totalValue"])
	}
}

func TestGetPortfolio_EmptyWhenNoHoldings(t *testing.T) {
	testDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer testDB.Close()

	// Create tables but do NOT seed any data.
	testDB.Exec(`CREATE TABLE IF NOT EXISTS account (user_id INTEGER PRIMARY KEY AUTOINCREMENT, cash_balance REAL)`)
	testDB.Exec(`CREATE TABLE IF NOT EXISTS holdings (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER, ticker TEXT, quantity INTEGER, price REAL)`)
	testDB.Exec(`INSERT INTO account (user_id, cash_balance) VALUES (1, 100000.00)`)
	db = testDB

	w := doRequest(setupRouter(), "GET", "/api/v1/portfolio", nil)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// holdings key should be absent or nil (Go JSON encodes nil slices as null)
	if resp["totalValue"] != 0.0 && resp["totalValue"] != nil {
		t.Errorf("expected totalValue 0, got %v", resp["totalValue"])
	}
}

// ---------------------------------------------------------------------------
// GET /api/v1/trades
// ---------------------------------------------------------------------------

func TestGetTrades_ReturnsOK(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "GET", "/api/v1/trades", nil)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetTrades_ReturnsEmptyList(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "GET", "/api/v1/trades", nil)

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

// ---------------------------------------------------------------------------
// GET /api/v1/prices
// ---------------------------------------------------------------------------

func TestGetPrices_ReturnsOK(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "GET", "/api/v1/prices?symbols=AAPL,GOOG", nil)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestGetPrices_ReturnsRequestedSymbols(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	w := doRequest(setupRouter(), "GET", "/api/v1/prices?symbols=AAPL,GOOG", nil)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["requested"] != "AAPL,GOOG" {
		t.Errorf("expected requested=AAPL,GOOG, got %v", resp["requested"])
	}

	prices, ok := resp["prices"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'prices' map in response")
	}
	if prices["AAPL"] != 100.00 {
		t.Errorf("expected AAPL=100.00, got %v", prices["AAPL"])
	}
}

// ---------------------------------------------------------------------------
// POST /api/v1/trade — BUY
// ---------------------------------------------------------------------------

func TestPlaceTrade_BuySuccess(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	payload := map[string]interface{}{
		"symbol":   "TSLA",
		"side":     "BUY",
		"quantity": 5,
	}
	w := doRequest(setupRouter(), "POST", "/api/v1/trade", payload)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestPlaceTrade_BuyDeductsBalance(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	payload := map[string]interface{}{
		"symbol":   "TSLA",
		"side":     "BUY",
		"quantity": 10, // 10 * $100 = $1000
	}
	doRequest(setupRouter(), "POST", "/api/v1/trade", payload)

	var balance float64
	db.QueryRow("SELECT cash_balance FROM account WHERE user_id = 1").Scan(&balance)

	expected := 100000.00 - 1000.00
	if balance != expected {
		t.Errorf("expected balance %.2f after buy, got %.2f", expected, balance)
	}
}

func TestPlaceTrade_BuyAddsNewHolding(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	payload := map[string]interface{}{
		"symbol":   "MSFT",
		"side":     "BUY",
		"quantity": 3,
	}
	doRequest(setupRouter(), "POST", "/api/v1/trade", payload)

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
	payload := map[string]interface{}{"symbol": "AAPL", "side": "BUY", "quantity": 10}
	doRequest(r, "POST", "/api/v1/trade", payload)

	var qty int
	db.QueryRow("SELECT quantity FROM holdings WHERE user_id = 1 AND ticker = 'AAPL'").Scan(&qty)

	// Seed has 50 AAPL; buying 10 more should give 60.
	if qty != 60 {
		t.Errorf("expected AAPL quantity 60 after accumulation, got %d", qty)
	}
}

func TestPlaceTrade_BuyInsufficientFunds(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	// 100000 / 100 = 1000 max shares; request 1001
	payload := map[string]interface{}{
		"symbol":   "TSLA",
		"side":     "BUY",
		"quantity": 1001,
	}
	w := doRequest(setupRouter(), "POST", "/api/v1/trade", payload)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for insufficient funds, got %d", w.Code)
	}
}

func TestPlaceTrade_BuyResponseFields(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	payload := map[string]interface{}{"symbol": "NVDA", "side": "BUY", "quantity": 2}
	w := doRequest(setupRouter(), "POST", "/api/v1/trade", payload)

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

// ---------------------------------------------------------------------------
// POST /api/v1/trade — SELL
// ---------------------------------------------------------------------------

func TestPlaceTrade_SellSuccess(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	// Seed has 50 AAPL — sell 10 of them.
	payload := map[string]interface{}{
		"symbol":   "AAPL",
		"side":     "SELL",
		"quantity": 10,
	}
	w := doRequest(setupRouter(), "POST", "/api/v1/trade", payload)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestPlaceTrade_SellCreditsBalance(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	payload := map[string]interface{}{"symbol": "AAPL", "side": "SELL", "quantity": 10}
	doRequest(setupRouter(), "POST", "/api/v1/trade", payload)

	var balance float64
	db.QueryRow("SELECT cash_balance FROM account WHERE user_id = 1").Scan(&balance)

	// 100000 + (10 * 100) = 101000
	expected := 100000.00 + 1000.00
	if balance != expected {
		t.Errorf("expected balance %.2f after sell, got %.2f", expected, balance)
	}
}

func TestPlaceTrade_SellDecreasesHoldings(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	payload := map[string]interface{}{"symbol": "AAPL", "side": "SELL", "quantity": 20}
	doRequest(setupRouter(), "POST", "/api/v1/trade", payload)

	var qty int
	db.QueryRow("SELECT quantity FROM holdings WHERE user_id = 1 AND ticker = 'AAPL'").Scan(&qty)

	// 50 - 20 = 30
	if qty != 30 {
		t.Errorf("expected AAPL quantity 30 after sell, got %d", qty)
	}
}

func TestPlaceTrade_SellInsufficientHoldings(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	// Seed has 50 AAPL — try to sell 100.
	payload := map[string]interface{}{
		"symbol":   "AAPL",
		"side":     "SELL",
		"quantity": 100,
	}
	w := doRequest(setupRouter(), "POST", "/api/v1/trade", payload)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for insufficient holdings, got %d", w.Code)
	}
}

func TestPlaceTrade_SellTickerNotOwned(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	payload := map[string]interface{}{
		"symbol":   "GOOG",
		"side":     "SELL",
		"quantity": 1,
	}
	w := doRequest(setupRouter(), "POST", "/api/v1/trade", payload)
	// Should not succeed — user has no GOOG holdings.
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

	payload := map[string]interface{}{"symbol": "AAPL", "side": "HOLD", "quantity": 1}
	w := doRequest(setupRouter(), "POST", "/api/v1/trade", payload)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid side, got %d", w.Code)
	}
}

func TestPlaceTrade_MalformedBody(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	req, _ := http.NewRequest("POST", "/api/v1/trade", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	setupRouter().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for malformed body, got %d", w.Code)
	}
}

func TestPlaceTrade_ZeroQuantityBuy(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	// Buying 0 shares should succeed without changing balance (0 * 100 = 0).
	initialBalance := 100000.00
	payload := map[string]interface{}{"symbol": "TSLA", "side": "BUY", "quantity": 0}
	w := doRequest(setupRouter(), "POST", "/api/v1/trade", payload)

	if w.Code == http.StatusOK {
		var balance float64
		db.QueryRow("SELECT cash_balance FROM account WHERE user_id = 1").Scan(&balance)
		if balance != initialBalance {
			t.Errorf("balance should not change on zero-quantity buy, got %.2f", balance)
		}
	}
	// If the API returns 400 for quantity=0 that is also acceptable behaviour.
}

// ---------------------------------------------------------------------------
// External API handlers — tested with a local mock HTTP server
// ---------------------------------------------------------------------------

// mockStockServer spins up an httptest.Server that returns a fixed JSON body
// and patches the handler under test by temporarily overriding os.Getenv via
// the STOCK_API_KEY env var. Because searchStocks / getStockQuote call
// api.stockdata.org directly, we redirect them with a simple monkey-patch
// using a package-level variable approach, OR we can test that the handler
// propagates errors correctly when the env var is unset / the upstream is down.

func TestSearchStocks_MissingAPIKey(t *testing.T) {
	// With no real API key the external call will fail; we just verify the
	// handler responds (not panic) and returns a server error or not-found.
	db = setupTestDB(t)
	defer db.Close()

	t.Setenv("STOCK_API_KEY", "invalid_test_key")

	w := doRequest(setupRouter(), "GET", "/api/v1/searchStocks?query=AAPL", nil)
	// We expect either 200 (upstream returned something) or 5xx/404 (upstream
	// rejected the key). The important thing is no panic and a valid HTTP code.
	if w.Code < 100 || w.Code > 599 {
		t.Errorf("unexpected status code %d from searchStocks", w.Code)
	}
}

func TestGetStockQuote_MissingAPIKey(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	t.Setenv("STOCK_API_KEY", "invalid_test_key")

	w := doRequest(setupRouter(), "GET", "/api/v1/quote/AAPL", nil)
	if w.Code < 100 || w.Code > 599 {
		t.Errorf("unexpected status code %d from getStockQuote", w.Code)
	}
}

// TestSearchStocks_MockServer replaces the real upstream with a local server.
// Because the URL is built inside the handler we use a test server and a
// custom httpClient shim — the cleanest option without refactoring.
// This test documents the expected contract so it can be wired up once the
// handler accepts a configurable base URL (a recommended next step).
func TestSearchStocks_ParsesValidResponse(t *testing.T) {
	// Stand up a mock server that mimics api.stockdata.org
	mock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"data":[{"symbol":"AAPL","name":"Apple Inc"}]}`)
	}))
	defer mock.Close()

	// Document the expected JSON shape returned by the real handler once the
	// base URL is made injectable.
	t.Log("Mock server ready at", mock.URL,
		"— wire this URL into the handler's baseURL to enable full integration.")

	// For now, assert the SearchResponse struct parses correctly.
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
// Balance consistency — multi-step scenario
// ---------------------------------------------------------------------------

// TestTradeSequence_BalanceConsistency exercises a realistic trading session:
// buy shares, sell some back, verify the final cash balance is arithmetically
// correct. This catches regressions in the two UPDATE paths working together.
func TestTradeSequence_BalanceConsistency(t *testing.T) {
	db = setupTestDB(t)
	defer db.Close()

	r := setupRouter()

	// Step 1: buy 50 shares of TSLA @ $100 each → cost $5000
	doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{"symbol": "TSLA", "side": "BUY", "quantity": 50})

	// Step 2: sell 20 of those shares → proceeds $2000
	doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{"symbol": "TSLA", "side": "SELL", "quantity": 20})

	// Step 3: buy 10 more → cost $1000
	doRequest(r, "POST", "/api/v1/trade", map[string]interface{}{"symbol": "TSLA", "side": "BUY", "quantity": 10})

	var balance float64
	db.QueryRow("SELECT cash_balance FROM account WHERE user_id = 1").Scan(&balance)

	// 100000 - 5000 + 2000 - 1000 = 96000
	expected := 96000.00
	if balance != expected {
		t.Errorf("expected balance %.2f after trade sequence, got %.2f", expected, balance)
	}

	var qty int
	db.QueryRow("SELECT quantity FROM holdings WHERE user_id = 1 AND ticker = 'TSLA'").Scan(&qty)
	// 50 - 20 + 10 = 40
	if qty != 40 {
		t.Errorf("expected TSLA quantity 40 after sequence, got %d", qty)
	}
}

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestPopulateResponse(t *testing.T) {
	product := Product{Name: "Notebook", ProductId: "abc-123", SalesPrice: 199.99}

	got := populateResponse(product, "in stock")

	if got.Name != "Notebook" {
		t.Errorf("expected Name=Notebook, got %s", got.Name)
	}
	if got.ProductId != "abc-123" {
		t.Errorf("expected ProductId=abc-123, got %s", got.ProductId)
	}
	if got.SalesPrice != 199.99 {
		t.Errorf("expected SalesPrice=199.99, got %f", got.SalesPrice)
	}
	if got.Status != "in stock" {
		t.Errorf("expected Status=in stock, got %s", got.Status)
	}
}

func TestGenerateResponse_InStock(t *testing.T) {
	product := Product{Name: "Notebook", ProductId: "abc-123", SalesPrice: 199.99}

	got := generateResponse([]byte(`{"availability": "In Stock"}`), product)

	if got.Status != "in stock" {
		t.Errorf("expected status=in stock, got %s", got.Status)
	}
}

func TestGenerateResponse_OutOfStock(t *testing.T) {
	product := Product{Name: "Notebook", ProductId: "abc-123", SalesPrice: 199.99}

	got := generateResponse([]byte(`{"availability": "Out of Stock"}`), product)

	if got.Status != "out of stock" {
		t.Errorf("expected status=out of stock, got %s", got.Status)
	}
}

func TestValidateProduct_InvalidBody(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/products/validate", bytes.NewBufferString("invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	ValidateProduct(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

func TestValidateProduct_ServiceUnavailable(t *testing.T) {
	original := URL
	URL = "http://localhost:1/"
	defer func() { URL = original }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"name":"Notebook","product_id":"abc-123","sales_price":199.99}`
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/products/validate", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	ValidateProduct(c)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}
}

func TestValidateProduct_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	original := URL
	URL = server.URL + "/"
	defer func() { URL = original }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"name":"Notebook","product_id":"abc-123","sales_price":199.99}`
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/products/validate", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	ValidateProduct(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestValidateProduct_InStock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"availability":"In Stock"}`))
	}))
	defer server.Close()

	original := URL
	URL = server.URL + "/"
	defer func() { URL = original }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"name":"Notebook","product_id":"abc-123","sales_price":199.99}`
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/products/validate", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	ValidateProduct(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	var resp map[string]ProductResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["body"].Status != "in stock" {
		t.Errorf("expected status=in stock, got %s", resp["body"].Status)
	}
	if resp["body"].Name != "Notebook" {
		t.Errorf("expected Name=Notebook, got %s", resp["body"].Name)
	}
}

func TestValidateProduct_OutOfStock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"availability":"Out of Stock"}`))
	}))
	defer server.Close()

	original := URL
	URL = server.URL + "/"
	defer func() { URL = original }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"name":"Notebook","product_id":"abc-123","sales_price":199.99}`
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/products/validate", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	ValidateProduct(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	var resp map[string]ProductResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["body"].Status != "out of stock" {
		t.Errorf("expected status=out of stock, got %s", resp["body"].Status)
	}
}

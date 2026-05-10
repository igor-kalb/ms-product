package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

var URL = os.Getenv("WIREMOCK_URL")

type Product struct {
	Name       string  `json:"name"`
	ProductId  string  `json:"product_id"`
	SalesPrice float32 `json:"sales_price"`
}

type ProductResponse struct {
	Name       string  `json:"name"`
	ProductId  string  `json:"product_id"`
	SalesPrice float32 `json:"sales_price"`
	Status     string  `json:"status"`
}

func ValidateProduct(c *gin.Context) {
	var json Product
	if !validatedRequest(c, &json) {
		return
	}

	var url = URL + json.ProductId
	resp, err := http.Get(url)

	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"err": err})
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		c.JSON(http.StatusNotFound, gin.H{"err": err, "message": "NOT FOUND"})
		return
	}

	data, _ := io.ReadAll(resp.Body)

	jsonResponse := generateResponse(data, json)

	c.JSON(200, gin.H{"body": *jsonResponse})
}

func generateResponse(data []byte, json Product) *ProductResponse {
	jsonResponse := new(ProductResponse)

	if strings.Contains(string(data), "In Stock") {
		jsonResponse = populateResponse(json, "in stock")
	} else {
		jsonResponse = populateResponse(json, "out of stock")
	}
	return jsonResponse
}

func validatedRequest(c *gin.Context, json *Product) bool {
	if err := c.ShouldBindJSON(json); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return false
	}
	return true
}

func populateResponse(
	json Product, status string) *ProductResponse {

	response := new(ProductResponse)

	response.Name = json.Name
	response.ProductId = json.ProductId
	response.SalesPrice = json.SalesPrice
	response.Status = status

	return response
}

func main() {
	fmt.Println("Hello World")
	r := gin.Default()

	v1 := r.Group("/v1")
	{
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
			})
		})

		v1.POST("products/validate", ValidateProduct)
	}

	r.Run(":8000")
}

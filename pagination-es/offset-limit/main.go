package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/gin-gonic/gin"
)

type Product struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
}

var es *elasticsearch.Client

func main() {
	// Initialize Elasticsearch client
	var err error
	es, err = elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	// Index sample data
	indexSampleData()

	// Set up Gin router
	r := gin.Default()
	r.GET("/products", getProducts)
	r.Run()
}

func indexSampleData() {
	products := []Product{
		{ID: 1, Name: "Product 1", Price: 100},
		{ID: 2, Name: "Product 2", Price: 200},
		{ID: 3, Name: "Product 3", Price: 300},
		{ID: 4, Name: "Product 4", Price: 400},
		{ID: 5, Name: "Product 5", Price: 500},
		{ID: 6, Name: "Product 6", Price: 600},
		{ID: 7, Name: "Product 7", Price: 700},
		{ID: 8, Name: "Product 8", Price: 800},
		{ID: 9, Name: "Product 9", Price: 900},
		{ID: 10, Name: "Product 10", Price: 1000},
	}

	for _, product := range products {
		data, _ := json.Marshal(product)
		req := esapi.IndexRequest{
			Index:      "products",
			DocumentID: strconv.Itoa(product.ID),
			Body:       bytes.NewReader(data),
			Refresh:    "true",
		}
		res, err := req.Do(context.Background(), es)
		if err != nil {
			log.Fatalf("Error getting response: %s", err)
		}
		defer res.Body.Close()
		if res.IsError() {
			log.Printf("[%s] Error indexing document ID=%d", res.Status(), product.ID)
		} else {
			log.Printf("[%s] Document ID=%d indexed.", res.Status(), product.ID)
		}
	}
}

func getProducts(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	from := (page - 1) * limit // ~offset

	var buf bytes.Buffer
	query := map[string]interface{}{
		"from": from,
		"size": limit,
		"query": map[string]interface{}{
			"match_all": map[string]interface{}{},
		},
	}
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	res, err := es.Search(
		es.Search.WithContext(context.Background()),
		es.Search.WithIndex("products"),
		es.Search.WithBody(&buf),
		es.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			log.Fatalf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	}

	var r map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	}

	products := make([]Product, 0)
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		var product Product
		source := hit.(map[string]interface{})["_source"]
		data, _ := json.Marshal(source)
		json.Unmarshal(data, &product)
		products = append(products, product)
	}

	c.JSON(200, products)
}

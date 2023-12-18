package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"gofr.dev/pkg/gofr"
)

// Customer struct to represent a customer
type Customer struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Config struct to hold database configuration
type Config struct {
	DB struct {
		Driver   string `json:"Driver"`
		User     string `json:"User"`
		Password string `json:"Password"`
		Host     string `json:"Host"`
		Port     string `json:"Port"`
		Database string `json:"Database"`
	} `json:"DB"`
}

func main() {
	// Load database configuration from the config file
	dbConfig, err := loadConfig("configs/config.json")
	if err != nil {
		log.Fatal("Error loading configuration:", err)
	}

	// Initialize a database connection
	db, err := sql.Open(dbConfig.DB.Driver, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		dbConfig.DB.User, dbConfig.DB.Password, dbConfig.DB.Host, dbConfig.DB.Port, dbConfig.DB.Database))
	if err != nil {
		log.Fatal("Error opening database connection:", err)
	}
	defer db.Close()

	// initialise gofr object
	app := gofr.New()

	app.GET("/greet", func(ctx *gofr.Context) (interface{}, error) {
		// Get the value using the redis instance
		value, err := ctx.Redis.Get(ctx.Context, "greeting").Result()

		return value, err
	})

	app.POST("/customer/{name}", func(ctx *gofr.Context) (interface{}, error) {
		name := ctx.PathParam("name")

		// Inserting a customer row in the database using SQL
		_, err := db.ExecContext(ctx.Context, "INSERT INTO customers (name) VALUES (?)", name)

		return nil, err
	})

	app.GET("/customer", func(ctx *gofr.Context) (interface{}, error) {
		var customers []Customer

		// Getting the customer from the database using SQL
		rows, err := db.QueryContext(ctx.Context, "SELECT * FROM customers")
		if err != nil {
			return nil, err
		}

		for rows.Next() {
			var customer Customer
			if err := rows.Scan(&customer.ID, &customer.Name); err != nil {
				return nil, err
			}

			customers = append(customers, customer)
		}

		// return the customer
		return customers, nil
	})

	// Starts the server, it will listen on the default port 8000.
	// it can be over-ridden through configs
	app.Start()
}

func loadConfig(filename string) (Config, error) {
	var config Config
	configFile, err := os.Open(filename)
	if err != nil {
		return config, err
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	return config, err
}

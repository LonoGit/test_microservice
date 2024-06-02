package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	ServicePort string

	DBName string
	DBHost string
	DBPort string

	DBUser     string
	DBPassword string
}

type Log struct {
	ID      uint      `gorm:"primaryKey"`
	Time    time.Time `json:"time"`
	Service string    `json:"service"`
	Message string    `json:"message"`
}

var db *gorm.DB
var config Config

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("No .env file found")
	}

	config = Config{
		DBUser:      os.Getenv("DB_USER"),
		DBPassword:  os.Getenv("DB_PASSWORD"),
		DBName:      os.Getenv("DB_NAME"),
		DBHost:      os.Getenv("DB_HOST"),
		DBPort:      os.Getenv("DB_PORT"),
		ServicePort: os.Getenv("SERVICE_PORT"),
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", config.DBUser, config.DBPassword, config.DBHost, config.DBPort, config.DBName)

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database")
	}

	db.AutoMigrate(&Log{})
}

func main() {
	router := gin.Default()

	router.POST("/logs", createLog)
	router.GET("/logs", getLogs)

	router.Run(fmt.Sprintf(":%s", config.ServicePort))
}

func createLog(c *gin.Context) {
	var log Log
	if err := c.ShouldBindJSON(&log); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Time = time.Now()
	db.Create(&log)
	c.JSON(http.StatusOK, log)
}

func getLogs(c *gin.Context) {
	var logs []Log
	if result := db.Find(&logs); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

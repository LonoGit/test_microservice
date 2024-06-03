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
	DBURL  string

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

const (
	maxConnectionAttempts = 5
	timeout               = 2 * time.Second
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error of reading variables of environment")
	}

	config = Config{
		DBUser:      os.Getenv("DB_USER"),
		DBPassword:  os.Getenv("DB_PASSWORD"),
		DBName:      os.Getenv("DB_NAME"),
		DBHost:      os.Getenv("DB_HOST"),
		DBPort:      os.Getenv("DB_PORT"),
		ServicePort: os.Getenv("SERVICE_PORT"),
	}

	config.DBURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", config.DBUser, config.DBPassword, config.DBHost, config.DBPort, config.DBName)

	if err := checkDBConnection(); err != nil {
		log.Println("Failed to connect to the database, the service started without connecting to the database")
		return
	}

	migrateDB()
}

func main() {
	router := gin.Default()

	router.POST("/logs", withDBConnection(createLog))
	router.GET("/logs", withDBConnection(getLogs))

	router.Run(fmt.Sprintf(":%s", config.ServicePort))
}

func withDBConnection(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		if err = checkDBConnection(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
			return
		}

		migrateDB()
		handler(c)
	}
}

func checkDBConnection() error {
	var err error
	for i := 0; i < maxConnectionAttempts; i++ {
		db, err = gorm.Open(postgres.Open(config.DBURL), &gorm.Config{})
		if err == nil {
			return nil
		}
		time.Sleep(timeout)
	}

	return err
}

func migrateDB() error {
	return db.AutoMigrate(&Log{})
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

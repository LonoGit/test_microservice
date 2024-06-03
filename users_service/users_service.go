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

type User struct {
	ID    uint   `gorm:"primaryKey" json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
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

	router.GET("/users", withDBConnection(getUsers))
	router.GET("/users/:id", withDBConnection(getUserByID))
	router.POST("/users", withDBConnection(createUser))
	router.PUT("/users/:id", withDBConnection(updateUser))
	router.DELETE("/users/:id", withDBConnection(deleteUser))

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
	return db.AutoMigrate(&User{})
}

func createUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.Create(&user)
	c.JSON(http.StatusOK, user)
}

func getUsers(c *gin.Context) {
	var users []User
	if result := db.Find(&users); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func getUserByID(c *gin.Context) {
	id := c.Param("id")
	var user User
	if result := db.First(&user, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func updateUser(c *gin.Context) {
	id := c.Param("id")
	var user User
	if result := db.First(&user, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db.Save(&user)
	c.JSON(http.StatusOK, user)
}

func deleteUser(c *gin.Context) {
	id := c.Param("id")
	result := db.Delete(&User{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

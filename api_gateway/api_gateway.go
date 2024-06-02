package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Config struct {
	ServicePort string

	UsersServicePort string
	LogServicePort   string

	UsersServiceHost string
	LogServiceHost   string

	UsersServiceURL string
	LogServiceURL   string
}

var config Config

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("Error of reading variables of environment")
	}

	config = Config{
		ServicePort:      os.Getenv("SERVICE_PORT"),
		UsersServicePort: os.Getenv("USERS_SERVICE_PORT"),
		LogServicePort:   os.Getenv("LOG_SERVICE_PORT"),
		UsersServiceHost: os.Getenv("USERS_SERVICE_HOST"),
		LogServiceHost:   os.Getenv("LOG_SERVICE_HOST"),
	}

	config.UsersServiceURL = fmt.Sprintf("http://%s:%s", config.UsersServiceHost, config.UsersServicePort)
	config.LogServiceURL = fmt.Sprintf("http://%s:%s", config.LogServiceHost, config.LogServicePort)
}

func main() {
	router := gin.Default()

	router.GET("/service", proxyRequest(fmt.Sprintf("%s/users", config.UsersServiceURL)))
	router.GET("/service/:id", proxyRequest(fmt.Sprintf("%s/users/:id", config.UsersServiceURL)))
	router.POST("/service", proxyRequest(fmt.Sprintf("%s/users", config.UsersServiceURL)))
	router.PUT("/service/:id", proxyRequest(fmt.Sprintf("%s/users/:id", config.UsersServiceURL)))
	router.DELETE("/service/:id", proxyRequest(fmt.Sprintf("%s/users/:id", config.UsersServiceURL)))

	router.GET("/logs", proxyRequest(fmt.Sprintf("%s/logs", config.LogServiceURL)))

	router.Run(fmt.Sprintf(":%s", config.ServicePort))
}

func proxyRequest(url string) gin.HandlerFunc {
	return func(c *gin.Context) {
		proxyURL := url
		for _, param := range c.Params {
			proxyURL = strings.ReplaceAll(proxyURL, ":"+param.Key, param.Value)
		}

		req, err := http.NewRequest(c.Request.Method, proxyURL, c.Request.Body)
		if err != nil {
			log.Println(err)
		}
		req.Header = c.Request.Header

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
		}

		// Record logs in the service
		Request := fmt.Sprintf("Request: %s %s", c.Request.Method, c.FullPath())
		Response := fmt.Sprintf("Response: %s", strconv.Itoa(resp.StatusCode))
		logToService(Request, Response)

		c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
	}
}

func logToService(service, message string) {
	logEntry := map[string]string{
		"service": service,
		"message": message,
	}
	jsonData, _ := json.Marshal(logEntry)

	_, err := http.Post(fmt.Sprintf("%s/logs", config.LogServiceURL), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Failed to send log: %v", err)
	}
}

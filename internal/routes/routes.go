package routes

import (
	"net/http"

	"github.com/amir-saatchi/rest-api/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NewRouter initializes the Gin router with both DB instances
func NewRouter(mainDB *gorm.DB, secondaryDB *gorm.DB) *gin.Engine {
    router := gin.Default()

    // Pass both DB instances to handlers via context
    router.Use(func(c *gin.Context) {
        c.Set("mainDB", mainDB)
        c.Set("secondaryDB", secondaryDB)
        c.Next()
    })

    // Define routes
    router.GET("/", indexHandler)
    router.GET("/api/data", apiDataHandler)
    router.POST("/api/post", postHandler)

    router.GET("/users", getAllUsers)
    router.POST("/users", createUser)

    return router
}

func getAllUsers(c *gin.Context) {
    mainDB := c.MustGet("mainDB").(*gorm.DB)
    var users []models.User
    mainDB.Find(&users)
    c.JSON(200, users)
}

func createLog(c *gin.Context) {
    var logEntry models.Log
    if err := c.BindJSON(&logEntry); err != nil {
        c.AbortWithStatusJSON(400, gin.H{"error": "Invalid request"})
        return
    }

    secondaryDB := c.MustGet("secondaryDB").(*gorm.DB)
    result := secondaryDB.Create(&logEntry)
    if result.Error != nil {
        c.AbortWithStatusJSON(500, gin.H{"error": "Failed to create log"})
        return
    }

    c.JSON(201, logEntry)
}

func createUser(c *gin.Context) {
    var user models.User
    if err := c.BindJSON(&user); err != nil {
        c.AbortWithStatusJSON(400, gin.H{"error": "Invalid request"})
        return
    }

    dbInstance := c.MustGet("db").(*gorm.DB)
    result := dbInstance.Create(&user)
    if result.Error != nil {
        c.AbortWithStatusJSON(500, gin.H{"error": "Failed to create user"})
        return
    }

    c.JSON(201, user)
}

// Handler functions adapted for Gin's context
func indexHandler(c *gin.Context) {
    c.String(http.StatusOK, "Hello, world!")
}

func apiDataHandler(c *gin.Context) {
    data := map[string]string{
        "message": "API data here!",
    }
    c.JSON(http.StatusOK, data)
}

func postHandler(c *gin.Context) {
    var payload struct {
        Name string `json:"name"`
    }

    // Use Gin's BindJSON for easier JSON decoding
    if err := c.BindJSON(&payload); err != nil {
        c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Received name: " + payload.Name})
}


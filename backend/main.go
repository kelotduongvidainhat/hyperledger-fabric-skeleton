package main

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kelotduongvidainhat/hyperledger-fabric-skeleton/backend/internal/fabric"
)

var fabricClient *fabric.FabricClient

func main() {
	var err error
	// Initialize Fabric Client
	fabricClient, err = fabric.NewFabricClient()
	if err != nil {
		fmt.Printf("Error initializing Fabric Client: %s\n", err)
		return
	}
	defer fabricClient.Close()

	// Setup Gin router
	r := gin.Default()

	// Configure CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "X-User-ID"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Routes
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	r.GET("/assets", getAllAssets)
	r.GET("/assets/:id", getAsset)
	r.POST("/assets", createAsset)
	r.PUT("/assets/:id/transfer", transferAsset)

	// Admin Routes
	r.GET("/admin/identities", listIdentities)

	// Start server
	fmt.Println("Backend API starting on :8080...")
	r.Run(":8080")
}

func getAllAssets(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	assets, err := fabricClient.GetAllAssets(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, assets)
}

func getAsset(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	id := c.Param("id")
	asset, err := fabricClient.ReadAsset(userID, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, asset)
}

func createAsset(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	var req struct {
		ID             string `json:"id" binding:"required"`
		Color          string `json:"color" binding:"required"`
		Size           int    `json:"size" binding:"required"`
		Owner          string `json:"owner" binding:"required"`
		AppraisedValue int    `json:"appraisedValue" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := fabricClient.CreateAsset(userID, req.ID, req.Color, req.Size, req.Owner, req.AppraisedValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Asset created successfully"})
}

func transferAsset(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	id := c.Param("id")
	var req struct {
		NewOwner string `json:"newOwner" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := fabricClient.TransferAsset(userID, id, req.NewOwner)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Asset transferred successfully"})
}


func listIdentities(c *gin.Context) {
	// Simple mock for now, or read from file system
	identities := []string{"admin", "user1", "user2"}
	c.JSON(http.StatusOK, identities)
}

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kelotduongvidainhat/hyperledger-fabric-skeleton/backend/internal/fabric"
	"github.com/kelotduongvidainhat/hyperledger-fabric-skeleton/backend/internal/listener"
	_ "github.com/lib/pq"
)

var (
	fabricClient *fabric.FabricClient
	db           *sql.DB
)

func main() {
	var err error
	// 1. Initialize Fabric Client
	fabricClient, err = fabric.NewFabricClient()
	if err != nil {
		log.Fatalf("Error initializing Fabric Client: %v", err)
	}
	defer fabricClient.Close()

	// 2. Initialize Database
	connStr := "postgres://postgres:password@localhost:5432/fabric_assets?sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatalf("Error pinging database: %v", err)
	}
	fmt.Println("Connected to PostgreSQL database")

	// 3. Start Event Listener
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := listener.StartEventListener(ctx, fabricClient.Network, db); err != nil {
			log.Printf("Event listener error: %v", err)
		}
	}()

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

	// DB Query Routes
	r.GET("/api/query/assets", queryAssetsFromDB)

	// Start server
	fmt.Println("Backend API starting on :8080...")
	r.Run(":8080")
}

func queryAssetsFromDB(c *gin.Context) {
	rows, err := db.Query("SELECT id, color, size, owner, appraised_value FROM assets")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var assets []gin.H
	for rows.Next() {
		var id, color, owner string
		var size, value int
		if err := rows.Scan(&id, &color, &size, &owner, &value); err != nil {
			continue
		}
		assets = append(assets, gin.H{
			"ID":             id,
			"Color":          color,
			"Size":           size,
			"Owner":          owner,
			"AppraisedValue": value,
		})
	}
	c.JSON(http.StatusOK, assets)
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

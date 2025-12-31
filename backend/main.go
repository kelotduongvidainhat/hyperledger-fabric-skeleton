package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

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
	connStr := "postgres://postgres:password@localhost:5432/fabricdb?sslmode=disable"
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
	r.PUT("/assets/:id/lock", lockAsset)
	r.PUT("/assets/:id/unlock", unlockAsset)
	r.PUT("/assets/:id/transfer", transferAsset)
	r.GET("/assets/:id/history", getAssetHistory)

	// Admin Routes
	r.GET("/admin/identities", listIdentities)

	// Auth Routes
	r.POST("/auth/register", registerUser)
	r.POST("/auth/login", loginUser)

	// DB Query Routes
	r.GET("/api/query/assets", queryAssetsFromDB)

	// Start server
	fmt.Println("Backend API starting on :8080...")
	r.Run(":8080")
}

func queryAssetsFromDB(c *gin.Context) {
	rows, err := db.Query("SELECT id, name, category, status, updated, updated_by, owner FROM assets")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var assets []gin.H
	for rows.Next() {
		var id, name, category, status, updated, updatedBy, owner string
		if err := rows.Scan(&id, &name, &category, &status, &updated, &updatedBy, &owner); err != nil {
			continue
		}
		assets = append(assets, gin.H{
			"ID":             id,
			"Name":           name,
			"Category":       category,
			"Status":         status,
			"Updated":        updated,
			"UpdatedBy":      updatedBy,
			"Owner":          owner,
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
		ID       string `json:"id" binding:"required"`
		Name     string `json:"name" binding:"required"`
		Category string `json:"category" binding:"required"`
		Owner    string `json:"owner" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := fabricClient.CreateAsset(userID, req.ID, req.Name, req.Category, req.Owner)
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

func getAssetHistory(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	id := c.Param("id")
	history, err := fabricClient.GetAssetHistory(userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	// history is a JSON string, we need to unmarshal it to return valid JSON
	// or we can just parse it into interface{} and return it directly.
	// Since Fabric returns a JSON string, if we send it as string in JSON, it will be double escaped.
	// But let's rely on JSON raw message if we want, or just let client handle it string.
	// Actually, client.go returns string. Let's send it as raw JSON.
	
	// log the history for debugging
	log.Printf("Asset History for ID %s: %s", id, history)

	c.Header("Content-Type", "application/json")
	c.String(http.StatusOK, history)
}


func listIdentities(c *gin.Context) {
	usersDir := "/home/qwe/hyperledger-fabric-skeleton/network/crypto-config/peerOrganizations/org1.example.com/users"
	files, err := ioutil.ReadDir(usersDir)
	if err != nil {
		log.Printf("Error reading users dir: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list identities"})
		return
	}

	var identities []string
	for _, f := range files {
		if f.IsDir() {
			// Folder name is like "user1@org1.example.com"
			// Extract username part
			name := f.Name()
			parts := strings.Split(name, "@")
			if len(parts) > 0 {
				identities = append(identities, parts[0])
			}
		}
	}
	c.JSON(http.StatusOK, identities)
}
func lockAsset(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	id := c.Param("id")

	err := fabricClient.LockAsset(userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Asset locked successfully"})
}

func unlockAsset(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	id := c.Param("id")

	err := fabricClient.UnlockAsset(userID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Asset unlocked successfully"})
}

func registerUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Register the user using the CA Client
	// We use the hardcoded admin credentials for the registrar for this demo
	// In a real app, this might be a specific registrar identity
	secret, err := fabricClient.CAClient.Register("admin", "adminpw", req.Username, req.Password, "client", "org1.department1")
	if err != nil {
		// If user already exists, we might want to handle gracefully, but error is fine for now
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to register user: %v", err)})
		return
	}

	// Also create the user in the chaincode (optional, but good for keeping state)
	// We use Admin to create the user record on-chain
	err = fabricClient.CreateUser("admin", req.Username, req.Username, "USER")
	if err != nil {
		log.Printf("Warning: Failed to create user on chain: %v", err)
		// Don't fail the request, as CA registration succeeded
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully", 
		"secret": secret,
	})
}

func loginUser(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enroll the user (verify credentials and download certs to local wallet)
	_, err := fabricClient.CAClient.Enroll(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication failed or user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": req.Username,
	})
}


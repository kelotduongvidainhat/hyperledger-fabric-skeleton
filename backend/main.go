package main

import (
	"backend/internal/api"
	"backend/internal/auth"
	"backend/internal/fabric"
	"backend/internal/db"
	"backend/internal/models"
	"backend/internal/storage"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/hyperledger/fabric-gateway/pkg/client"
	shell "github.com/ipfs/go-ipfs-api"
)

func main() {
	// 1. Setup DB
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	database, err := db.InitDB(dbHost, dbPort, "admin", "adminpw", "ownership_registry")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// 2. CONFIGURATION
	cryptoPathOrg1 := os.Getenv("CRYPTO_PATH_ORG1")
	if cryptoPathOrg1 == "" {
		cryptoPathOrg1 = "../network/crypto-config/peerOrganizations/org1.example.com"
	}
	cryptoPathOrg2 := os.Getenv("CRYPTO_PATH_ORG2")
	if cryptoPathOrg2 == "" {
		cryptoPathOrg2 = "../network/crypto-config/peerOrganizations/org2.example.com"
	}
	
	// Org1 Admin paths
	certPath1 := fmt.Sprintf("%s/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem", cryptoPathOrg1)
	keyPath1 := ""
	tlsCertPath1 := fmt.Sprintf("%s/peers/peer0.org1.example.com/tls/ca.crt", cryptoPathOrg1)
	
	peerEndpoint := os.Getenv("PEER_ENDPOINT")
	if peerEndpoint == "" {
		peerEndpoint = "localhost:7051"
	}
	walletPath := os.Getenv("WALLET_PATH")
	if walletPath == "" {
		walletPath = "./wallet"
	}

	// Find Key File (Org1 Admin)
	keyDir1 := fmt.Sprintf("%s/users/Admin@org1.example.com/msp/keystore", cryptoPathOrg1)
	files, err := os.ReadDir(keyDir1)
	if err == nil {
		for _, file := range files {
			if file.Name() != "" {
				keyPath1 = fmt.Sprintf("%s/%s", keyDir1, file.Name())
				break
			}
		}
	}

	cfg := fabric.Config{
		CertPath:      certPath1,
		KeyPath:       keyPath1,
		TlsCertPath:   tlsCertPath1,
		PeerEndpoint:  peerEndpoint,
		GatewayPeer:   "peer0.org1.example.com",
		ChannelName:   "mychannel",
		ChaincodeName: "basic",
	}

	// 1. Establish gRPC Connection
	conn, err := fabric.CreateGRPCConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to create gRPC connection: %v", err)
	}
	defer conn.Close()

	// 2. Setup Auth Handler
	// CA URL is usually localhost:7054 for Org1 CA
	// TLS is disabled.
	ca1URL := os.Getenv("CA1_URL")
	if ca1URL == "" {
		ca1URL = "https://localhost:7054"
	}
	ca2URL := os.Getenv("CA2_URL")
	if ca2URL == "" {
		ca2URL = "https://localhost:8054"
	}

	caCfg1 := fabric.CAConfig{
		URL:           ca1URL,
		MSPID:         "Org1MSP",
		WalletPath:    walletPath,
		AdminPath:     cryptoPathOrg1 + "/users/Admin@org1.example.com/msp",
		CAName:        "ca-org1",
		ContainerName: "ca_org1",
	}
	caCfg2 := fabric.CAConfig{
		URL:           ca2URL,
		MSPID:         "Org2MSP",
		WalletPath:    walletPath,
		AdminPath:     cryptoPathOrg2 + "/users/Admin@org2.example.com/msp",
		CAName:        "ca-org2",
		ContainerName: "ca_org2",
	}
	authHandler := &api.AuthHandler{
		CAConfigs: map[string]fabric.CAConfig{
			"Org1MSP": caCfg1,
			"Org2MSP": caCfg2,
		},
		DB: database,
	}
	adminHandler := &api.AdminHandler{
		CAConfigs: []fabric.CAConfig{
			caCfg1,
			caCfg2,
		},
		WalletPath: walletPath,
		Config:     cfg,
		Conn:       conn,
		DB:         database,
	}

	// 4. Setup Storage Handler (IPFS + MinIO)
	ipfsURL := os.Getenv("IPFS_URL")
	if ipfsURL == "" {
		ipfsURL = "localhost:5001"
	}
	sh := shell.NewShell(ipfsURL)

	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "localhost:9000"
	}
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	minioBucket := os.Getenv("MINIO_BUCKET")
	if minioBucket == "" {
		minioBucket = "assets"
	}
	minioUseSSL := os.Getenv("MINIO_USE_SSL") == "true"

	minioStore, err := storage.NewMinIOStorage(minioEndpoint, minioAccessKey, minioSecretKey, minioBucket, minioUseSSL)
	if err != nil {
		log.Printf("Warning: Failed to initialize MinIO: %v", err)
	} else {
		publicEndpoint := os.Getenv("MINIO_PUBLIC_ENDPOINT")
		if publicEndpoint != "" {
			err = minioStore.SetPublicEndpoint(publicEndpoint, minioAccessKey, minioSecretKey, minioUseSSL)
			if err != nil {
				log.Printf("Warning: Failed to set MinIO public endpoint: %v", err)
			}
		}
	}

	storageHandler := &api.StorageHandler{
		MinIO: minioStore,
		Ipfs:  sh,
	}

	// SETUP SERVER
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10 MB
	})
	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOrigins:     "http://localhost:5173", // Frontend Dev Port
	}))

	// 3. START EVENTUAL CONSISTENCY LISTENER
	// We use the Admin identity to listen for all events across the organization
	go func() {
		// Event listener usually runs as Org1 Admin
		adminID, adminSign, err := fabric.GetIdentity("admin", "Org1MSP", walletPath)
		if err != nil {
			log.Printf("Listener Error: Could not load admin identity: %v", err)
			return
		}
		
		gw, err := fabric.CreateGateway(conn, adminID, adminSign)
		if err != nil {
			log.Printf("Listener Error: Could not connect to gateway: %v", err)
			return
		}
		defer gw.Close()

		network := gw.GetNetwork(cfg.ChannelName)
		fabric.StartEventListener(context.Background(), network, database)
	}()

	// PUBLIC ROUTES
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Ownership Registry API Running")
	})
	app.Post("/auth/login", authHandler.Login)
	app.Post("/auth/register", authHandler.Register)
	app.Post("/auth/refresh", authHandler.Refresh)
	app.Post("/auth/logout", authHandler.Logout)
	app.Delete("/auth/me", auth.Middleware(), authHandler.DeleteAccount)

	// STORAGE ROUTES
	app.Post("/api/storage/upload", auth.Middleware(), storageHandler.Upload)
	app.Get("/api/storage/url/:objectName", auth.Middleware(), storageHandler.GetURL)

	// OPA MIDDLEWARE: Centralized AuthZ delegation
	app.Use(func(c *fiber.Ctx) error {
		// Only apply to protected sub-groups or API routes
		path := c.Path()
		if !strings.HasPrefix(path, "/admin") && !strings.HasPrefix(path, "/assets") && !strings.HasPrefix(path, "/api/ipfs") {
			return c.Next()
		}

		// Get user context (populated by auth.Middleware which should run before or inside groups)
		// Since auth.Middleware() is applied per group below, we need a way to ensure 
		// OPA runs AFTER JWT validation but before hander.
		return c.Next()
	})

	// ADMIN ROUTES (Protected + OPA Check)
	adminGroup := app.Group("/admin", auth.Middleware())
	adminGroup.Use(func(c *fiber.Ctx) error {
		user := c.Locals("user").(string)
		role := c.Locals("role").(string)
		org := c.Locals("org").(string)

		allowed, err := auth.CheckAuthorization(user, role, org, c.Method(), c.Path())
		if err != nil {
			log.Printf("OPA Error: %v", err)
			return c.Status(500).JSON(fiber.Map{"error": "Authorization service error"})
		}

		if !allowed {
			return c.Status(403).JSON(fiber.Map{"error": "Permission denied (OPA)"})
		}
		return c.Next()
	})
	adminGroup.Get("/stats", adminHandler.GetStats)
	adminGroup.Get("/users", adminHandler.GetUsers)
	adminGroup.Post("/users/:username/status", adminHandler.UpdateUserStatus)
	adminGroup.Get("/assets", adminHandler.GetAdminAssets)
	adminGroup.Post("/assets/:id/status", adminHandler.UpdateAssetStatus)
	adminGroup.Post("/sync", adminHandler.Sync)

	// PROTECTED ROUTES
	api := app.Group("/assets", auth.Middleware())

	// Helper to get Contract for the logged-in user
	getContract := func(c *fiber.Ctx) (*client.Gateway, *client.Contract, error) {
		username := c.Locals("user").(string)
		org := c.Locals("org").(string)
		
		// Load Identity from Wallet
		id, sign, err := fabric.GetIdentity(username, org, walletPath)
		if err != nil {
			return nil, nil, fmt.Errorf("identity not found for user %s (%s): %v", username, org, err)
		}

		// Create Gateway for this request
		gw, err := fabric.CreateGateway(conn, id, sign)
		if err != nil {
			return nil, nil, err
		}
		
		network := gw.GetNetwork(cfg.ChannelName)
		contract := network.GetContract(cfg.ChaincodeName)
		return gw, contract, nil
	}

	// Asset Routes (Protected)
	api.Get("/", func(c *fiber.Ctx) error {
		user := c.Locals("user").(string)
		role := c.Locals("role").(string)
		org := c.Locals("org").(string)

		// Consult OPA for high-level permission
		allowed, err := auth.CheckAuthorization(user, role, org, c.Method(), c.Path())
		if err != nil || !allowed {
			return c.Status(403).JSON(fiber.Map{"error": "Permission denied by OPA"})
		}

		// Non-admins: Filtered Database view (Data-level filtering still applies)
		if role != "admin" {
			fullID := fmt.Sprintf("%s::%s", org, user)

			var assets []models.Asset
			// Logic: Show if (PUBLIC OR Owner OR ProposedOwner) AND NOT DELETED
			err := database.Where("(UPPER(view) = 'PUBLIC' OR owner_id = ? OR proposed_owner_id = ?) AND status != 'DELETED'", fullID, fullID).Find(&assets).Error
			if err != nil {
				return c.Status(500).SendString("Database error: " + err.Error())
			}
			return c.JSON(assets)
		}

		// Admin gets blockchain
		gw, contract, err := getContract(c)
		if err != nil {
			return c.Status(401).SendString(err.Error())
		}
		defer gw.Close()

		result, err := contract.EvaluateTransaction("GetAllAssets")
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		var ledgerValues []models.LedgerValue
		if err := json.Unmarshal(result, &ledgerValues); err != nil {
			// Fallback
			return c.Type("json").Send(result)
		}

		var assets []models.Asset
		for _, val := range ledgerValues {
			asset := val.Asset
			asset.Action = val.Audit.Action
			asset.LastUpdatedBy = val.Audit.Actor
			if val.Audit.Timestamp != "" {
				t, err := time.Parse(time.RFC3339, val.Audit.Timestamp)
				if err == nil {
					asset.LastUpdatedAt = t
				}
			}
			assets = append(assets, asset)
		}
		return c.JSON(assets)
	})

	api.Post("/", func(c *fiber.Ctx) error {
		type CreateReq struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"desc"`
			ImageURL    string `json:"image_url"`
			ImageHash   string `json:"image_hash"`
			View        string `json:"view"`
			FileName    string `json:"file_name"`
			FileSize    int64  `json:"file_size"`
			FileHash    string `json:"file_hash"`
			IpfsCID     string `json:"ipfs_cid"`
			StoragePath string `json:"storage_path"`
			StorageType string `json:"storage_type"`
		}
		req := new(CreateReq)
		if err := c.BodyParser(req); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		gw, contract, err := getContract(c)
		if err != nil {
			return c.Status(401).SendString(err.Error())
		}
		defer gw.Close()


		_, err = contract.SubmitTransaction("CreateAsset", req.ID, req.Name, req.Description, req.ImageURL, req.ImageHash, req.View,
			req.FileName, fmt.Sprintf("%d", req.FileSize), req.FileHash, req.IpfsCID, req.StoragePath, req.StorageType)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.SendString("Asset Created")
	})

	api.Get("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")
		role := c.Locals("role").(string)

		// Non-admins: Filtered Database view
		if role != "admin" {
			username := c.Locals("user").(string)
			org := c.Locals("org").(string)
			fullID := fmt.Sprintf("%s::%s", org, username)

			var asset models.Asset
			if err := database.Where("id = ?", id).First(&asset).Error; err != nil {
				return c.Status(404).JSON(fiber.Map{"error": "Asset not found"})
			}

			// Privacy Check: Allow if Public, Owner, or Proposed Owner (Recipient)
			isPublic := strings.ToUpper(asset.View) == "PUBLIC"
			if !isPublic && asset.OwnerID != fullID && asset.ProposedOwnerID != fullID {
				return c.Status(403).JSON(fiber.Map{"error": "Private asset access denied"})
			}
			return c.JSON(asset)
		}

		gw, contract, err := getContract(c)
		if err != nil {
			return c.Status(401).SendString(err.Error())
		}
		defer gw.Close()

		result, err := contract.EvaluateTransaction("ReadAsset", id)
		if err != nil {
			return c.Status(404).SendString(err.Error())
		}

		var val models.LedgerValue
		if err := json.Unmarshal(result, &val); err != nil {
			return c.Type("json").Send(result)
		}

		asset := val.Asset
		asset.Action = val.Audit.Action
		asset.LastUpdatedBy = val.Audit.Actor
		if val.Audit.Timestamp != "" {
			t, err := time.Parse(time.RFC3339, val.Audit.Timestamp)
			if err == nil {
				asset.LastUpdatedAt = t
			}
		}

		return c.JSON(asset)
	})

	// Blockchain Verification Route (Direct Ledger access)
	api.Get("/:id/blockchain", func(c *fiber.Ctx) error {
		id := c.Params("id")
		role := c.Locals("role").(string)
		username := c.Locals("user").(string)
		org := c.Locals("org").(string)
		fullID := fmt.Sprintf("%s::%s", org, username)

		// 1. Permission Check (Same as Detail)
		if role != "admin" {
			var asset models.Asset
			if err := database.Where("id = ?", id).First(&asset).Error; err != nil {
				return c.Status(404).JSON(fiber.Map{"error": "Asset not found in off-chain database"})
			}
			isPublic := strings.ToUpper(asset.View) == "PUBLIC"
			if !isPublic && asset.OwnerID != fullID && asset.ProposedOwnerID != fullID {
				return c.Status(403).JSON(fiber.Map{"error": "Access to blockchain data restricted for this asset"})
			}
		}

		// 2. Fetch directly from Hyperledger Fabric
		gw, contract, err := getContract(c)
		if err != nil {
			return c.Status(401).SendString(err.Error())
		}
		defer gw.Close()

		result, err := contract.EvaluateTransaction("ReadAsset", id)
		if err != nil {
			return c.Status(404).SendString("Blockchain Read Error: " + err.Error())
		}

		// Return raw unmarshaled LedgerValue
		var val models.LedgerValue
		if err := json.Unmarshal(result, &val); err != nil {
			return c.Type("json").Send(result)
		}
		return c.JSON(val)
	})

	api.Get("/:id/history", func(c *fiber.Ctx) error {
		id := c.Params("id")
		role := c.Locals("role").(string)
		username := c.Locals("user").(string)
		org := c.Locals("org").(string)
		fullID := fmt.Sprintf("%s::%s", org, username)

		// Access check: Admin OR Owner
		if role != "admin" {
			var asset models.Asset
			if err := database.Where("id = ?", id).First(&asset).Error; err != nil {
				return c.Status(404).JSON(fiber.Map{"error": "Asset not found"})
			}
			if asset.OwnerID != fullID && asset.ProposedOwnerID != fullID {
				return c.Status(403).JSON(fiber.Map{"error": "Provenance history is restricted to the owner, proposed owner, or administrators."})
			}
		}

		gw, contract, err := getContract(c)
		if err != nil {
			return c.Status(401).SendString(err.Error())
		}
		defer gw.Close()

		result, err := contract.EvaluateTransaction("GetAssetHistory", id)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.Type("json").Send(result)
	})

	api.Post("/:id/view", func(c *fiber.Ctx) error {
		id := c.Params("id")
		type ViewReq struct {
			View string `json:"view"` // Public, Private
		}
		req := new(ViewReq)
		if err := c.BodyParser(req); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		gw, contract, err := getContract(c)
		if err != nil {
			return c.Status(401).SendString(err.Error())
		}
		defer gw.Close()

		_, err = contract.SubmitTransaction("UpdateAssetView", id, req.View)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		// Update in DB
		var asset models.Asset
		if err := database.Where("id = ?", id).First(&asset).Error; err == nil {
			asset.View = req.View
			database.Save(&asset)
		}

		return c.SendString("Asset Visibility Updated to " + req.View)
	})

	api.Post("/:id/transfer", func(c *fiber.Ctx) error {
		id := c.Params("id")
		type TransferReq struct {
			TargetUser string `json:"target_user"`
		}
		req := new(TransferReq)
		if err := c.BodyParser(req); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		// Automate Org::Username lookup
		var targetUser models.User
		searchQuery := req.TargetUser
		
		// If user provided Org::Username format, extract the username part for lookup
		if strings.Contains(req.TargetUser, "::") {
			parts := strings.Split(req.TargetUser, "::")
			searchQuery = parts[1]
		}

		if err := database.Where("username = ?", searchQuery).First(&targetUser).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("Target user %s not found in off-chain database", searchQuery)})
		}
		
		fullTargetID := fmt.Sprintf("%s::%s", targetUser.Org, targetUser.Username)

		gw, contract, err := getContract(c)
		if err != nil {
			return c.Status(401).SendString(err.Error())
		}
		defer gw.Close()

		_, err = contract.SubmitTransaction("ProposeTransfer", id, fullTargetID)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		// Create Notification for Target User
		senderUsername := c.Locals("user").(string)
		senderOrg := c.Locals("org").(string)
		fullSenderID := fmt.Sprintf("%s::%s", senderOrg, senderUsername)

		database.Create(&models.Notification{
			UserID:  fullTargetID,
			Title:   "Incoming Artifact Transfer",
			Message: fmt.Sprintf("%s has proposed an artifact transfer: %s", fullSenderID, id),
			Type:    "info",
			Link:    fmt.Sprintf("/assets/%s", id),
		})

		return c.SendString("Transfer Proposed to " + fullTargetID)
	})

	api.Post("/:id/accept", func(c *fiber.Ctx) error {
		id := c.Params("id")
		
		// Get Asset to know the current owner for notification
		var asset models.Asset
		database.Where("id = ?", id).First(&asset)
		oldOwner := asset.OwnerID

		gw, contract, err := getContract(c)
		if err != nil {
			return c.Status(401).SendString(err.Error())
		}
		defer gw.Close()

		_, err = contract.SubmitTransaction("AcceptTransfer", id)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		// Create Notification for Previous Owner
		currentUsername := c.Locals("user").(string)
		currentOrg := c.Locals("org").(string)
		fullCurrentID := fmt.Sprintf("%s::%s", currentOrg, currentUsername)

		database.Create(&models.Notification{
			UserID:  oldOwner,
			Title:   "Transfer Complete",
			Message: fmt.Sprintf("%s has accepted the transfer of %s", fullCurrentID, id),
			Type:    "success",
			Link:    fmt.Sprintf("/gallery/%s", id),
		})

		return c.SendString("Transfer Accepted")
	})

	api.Delete("/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		gw, contract, err := getContract(c)
		if err != nil {
			return c.Status(401).SendString(err.Error())
		}
		defer gw.Close()

		_, err = contract.SubmitTransaction("DeleteAsset", id)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		// Soft Delete in DB to maintain consistency
		var asset models.Asset
		if err := database.Where("id = ?", id).First(&asset).Error; err == nil {
			asset.Status = "DELETED"
			asset.View = "PRIVATE"
			database.Save(&asset)
		}

		return c.SendString("Asset Deleted")
	})

	// NOTIFICATIONS
	notifGroup := app.Group("/notifications", auth.Middleware())
	notifGroup.Get("/", func(c *fiber.Ctx) error {
		username := c.Locals("user").(string)
		org := c.Locals("org").(string)
		fullID := fmt.Sprintf("%s::%s", org, username)

		var notifications []models.Notification
		database.Where("user_id = ?", fullID).Order("created_at desc").Find(&notifications)
		return c.JSON(notifications)
	})

	notifGroup.Post("/:id/read", func(c *fiber.Ctx) error {
		id := c.Params("id")
		username := c.Locals("user").(string)
		org := c.Locals("org").(string)
		fullID := fmt.Sprintf("%s::%s", org, username)

		var notif models.Notification
		if err := database.Where("id = ? AND user_id = ?", id, fullID).First(&notif).Error; err != nil {
			return c.Status(404).SendString("Notification not found")
		}

		notif.IsRead = true
		database.Save(&notif)
		return c.SendString("Marked as read")
	})

	log.Println("Server running on port 3000")
	log.Fatal(app.Listen(":3000"))
}

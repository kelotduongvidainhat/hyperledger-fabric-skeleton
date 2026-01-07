package main

import (
	"backend/internal/api"
	"backend/internal/auth"
	"backend/internal/fabric"
	"backend/internal/db"
	"backend/internal/models"
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
	database, err := db.InitDB("localhost", "5432", "admin", "adminpw", "ownership_registry")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// 2. CONFIGURATION
	cryptoPath := "../network/crypto-config/peerOrganizations/org1.example.com"
	certPath := fmt.Sprintf("%s/users/Admin@org1.example.com/msp/signcerts/Admin@org1.example.com-cert.pem", cryptoPath)
	keyPath := fmt.Sprintf("%s/users/Admin@org1.example.com/msp/keystore/priv_sk", cryptoPath)
	tlsCertPath := fmt.Sprintf("%s/peers/peer0.org1.example.com/tls/ca.crt", cryptoPath)
	peerEndpoint := "localhost:7051"
	walletPath := "./wallet"

	// Find Key File (Admin)
	keyDir := fmt.Sprintf("%s/users/Admin@org1.example.com/msp/keystore", cryptoPath)
	files, err := os.ReadDir(keyDir)
	if err == nil {
		for _, file := range files {
			if file.Name() != "" {
				keyPath = fmt.Sprintf("%s/%s", keyDir, file.Name())
				break
			}
		}
	}

	cfg := fabric.Config{
		CertPath:      certPath,
		KeyPath:       keyPath,
		TlsCertPath:   tlsCertPath,
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
	caCfg1 := fabric.CAConfig{
		URL:           "https://localhost:7054",
		MSPID:         "Org1MSP",
		WalletPath:    walletPath,
		AdminPath:     cryptoPath + "/users/Admin@org1.example.com/msp",
		CAName:        "ca-org1",
		ContainerName: "ca_org1",
	}
	caCfg2 := fabric.CAConfig{
		URL:           "https://localhost:8054",
		MSPID:         "Org2MSP",
		WalletPath:    walletPath,
		AdminPath:     cryptoPath + "/users/Admin@org2.example.com/msp",
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

	// 3. Setup IPFS Handler
	sh := shell.NewShell("localhost:5001")
	ipfsHandler := &api.IPFSHandler{
		Shell: sh,
	}

	// SETUP SERVER
	app := fiber.New(fiber.Config{
		BodyLimit: 10 * 1024 * 1024, // 10 MB
	})
	app.Use(cors.New())

	// 3. START EVENTUAL CONSISTENCY LISTENER
	// We use the Admin identity to listen for all events across the organization
	go func() {
		adminID, adminSign, err := fabric.GetIdentity("admin", walletPath)
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
	app.Delete("/auth/me", auth.Middleware(), authHandler.DeleteAccount)

	// IPFS ROUTES
	app.Post("/api/ipfs/upload", auth.Middleware(), ipfsHandler.Upload)

	// ADMIN ROUTES (Protected + Role Check)
	adminGroup := app.Group("/admin", auth.Middleware())
	adminGroup.Use(func(c *fiber.Ctx) error {
		if c.Locals("role") != "admin" {
			return c.Status(403).JSON(fiber.Map{"error": "Admin access required"})
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
		
		// Load Identity from Wallet
		id, sign, err := fabric.GetIdentity(username, walletPath)
		if err != nil {
			return nil, nil, fmt.Errorf("identity not found for user %s: %v", username, err)
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
		role := c.Locals("role").(string)

		// Non-admins: Filtered Database view
		if role != "admin" {
			username := c.Locals("user").(string)
			org := c.Locals("org").(string)
			fullID := fmt.Sprintf("%s::%s", org, username)

			var assets []models.Asset
			// Logic: Show if (PUBLIC OR Owner) AND NOT DELETED
			err := database.Where("(UPPER(view) = 'PUBLIC' OR owner_id = ?) AND status != 'DELETED'", fullID).Find(&assets).Error
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

		_, err = contract.SubmitTransaction("CreateAsset", req.ID, req.Name, req.Description, req.ImageURL, req.ImageHash, req.View)
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

			// Privacy Check
			isPublic := strings.ToUpper(asset.View) == "PUBLIC"
			if !isPublic && asset.OwnerID != fullID {
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
			if asset.OwnerID != fullID {
				return c.Status(403).JSON(fiber.Map{"error": "Provenance history is restricted to the owner or administrators."})
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

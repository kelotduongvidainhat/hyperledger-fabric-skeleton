package main

import (
	"backend/internal/api"
	"backend/internal/auth"
	"backend/internal/fabric"
	"backend/internal/db"
	"backend/internal/models"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/hyperledger/fabric-gateway/pkg/client"
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

	// SETUP SERVER
	app := fiber.New()
	app.Use(cors.New())

	// PUBLIC ROUTES
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Ownership Registry API Running")
	})
	app.Post("/auth/login", authHandler.Login)
	app.Post("/auth/register", authHandler.Register)

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
		gw, contract, err := getContract(c)
		if err != nil {
			return c.Status(401).SendString(err.Error())
		}
		defer gw.Close()

		result, err := contract.EvaluateTransaction("GetAllAssets")
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.Type("json").Send(result)
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
		gw, contract, err := getContract(c)
		if err != nil {
			return c.Status(401).SendString(err.Error())
		}
		defer gw.Close()

		result, err := contract.EvaluateTransaction("ReadAsset", id)
		if err != nil {
			return c.Status(404).SendString(err.Error())
		}
		return c.Type("json").Send(result)
	})

	api.Get("/:id/history", func(c *fiber.Ctx) error {
		id := c.Params("id")
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
		if err := database.Where("username = ?", req.TargetUser).First(&targetUser).Error; err != nil {
			return c.Status(404).JSON(fiber.Map{"error": fmt.Sprintf("Target user %s not found in off-chain database", req.TargetUser)})
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
		return c.SendString(fmt.Sprintf("Transfer Proposed to %s", fullTargetID))
	})

	api.Post("/:id/accept", func(c *fiber.Ctx) error {
		id := c.Params("id")
		gw, contract, err := getContract(c)
		if err != nil {
			return c.Status(401).SendString(err.Error())
		}
		defer gw.Close()

		_, err = contract.SubmitTransaction("AcceptTransfer", id)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.SendString("Transfer Accepted")
	})

	log.Println("Server running on port 3000")
	log.Fatal(app.Listen(":3000"))
}

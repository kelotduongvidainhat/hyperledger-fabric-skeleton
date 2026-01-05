package main

import (
	"backend/internal/api"
	"backend/internal/auth"
	"backend/internal/fabric"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/hyperledger/fabric-gateway/pkg/client"
)

func main() {
	// CONFIGURATION
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
	// TLS is enabled, so use https. Insecure=true in ca.go handles self-signed certs.
	caCfg := fabric.CAConfig{
		URL:        "https://localhost:7054",
		MSPID:      "Org1MSP",
		WalletPath: walletPath,
		AdminPath:  cryptoPath + "/users/Admin@org1.example.com/msp",
	}
	authHandler := &api.AuthHandler{CAConfig: caCfg}

	// SETUP SERVER
	app := fiber.New()
	app.Use(cors.New())

	// PUBLIC ROUTES
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Ownership Registry API Running")
	})
	app.Post("/auth/login", authHandler.Login)
	app.Post("/auth/register", authHandler.Register)

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

		gw, contract, err := getContract(c)
		if err != nil {
			return c.Status(401).SendString(err.Error())
		}
		defer gw.Close()

		_, err = contract.SubmitTransaction("ProposeTransfer", id, req.TargetUser)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.SendString("Transfer Proposed")
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

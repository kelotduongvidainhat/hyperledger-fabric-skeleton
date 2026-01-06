package api

import (
	"backend/internal/fabric"
	"backend/internal/models"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type AdminHandler struct {
	CAConfig   fabric.CAConfig
	WalletPath string
	Config     fabric.Config
	Conn       interface{} // *grpc.ClientConn
	DB         *gorm.DB
}

type NetworkStats struct {
	TotalAssets      int `json:"total_assets"`
	TotalOwners      int `json:"total_owners"`
	PendingTransfers int `json:"pending_transfers"`
}

type IdentityInfo struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	DBStatus string `json:"db_status"`
	Email    string `json:"email"`
}

func (h *AdminHandler) GetStats(c *fiber.Ctx) error {
	// MOCK DATA for initial UI build
	stats := NetworkStats{
		TotalAssets:      42,
		TotalOwners:      12,
		PendingTransfers: 5,
	}

	return c.JSON(stats)
}

func (h *AdminHandler) GetUsers(c *fiber.Ctx) error {
	raw, err := fabric.ListIdentities(h.CAConfig)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	lines := strings.Split(strings.TrimSpace(raw), "\n")
	var identities []IdentityInfo

	// Regex to extract Name and Type
	reName := regexp.MustCompile(`Name: ([^,]+)`)
	reType := regexp.MustCompile(`Type: ([^,]+)`)

	for _, line := range lines {
		nameMatch := reName.FindStringSubmatch(line)
		typeMatch := reType.FindStringSubmatch(line)

		if len(nameMatch) > 1 && len(typeMatch) > 1 {
			name := nameMatch[1]
			identities = append(identities, IdentityInfo{
				Name:     name,
				Type:     typeMatch[1],
				DBStatus: "Synced",
				Email:    strings.ToLower(name) + "@example.org",
			})
		}
	}

	return c.JSON(fiber.Map{"identities": identities})
}

func (h *AdminHandler) GetAdminAssets(c *fiber.Ctx) error {
	source := c.Query("source", "blockchain")

	if source == "database" {
		var assets []models.Asset
		if err := h.DB.Find(&assets).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Database error: " + err.Error()})
		}
		return c.JSON(fiber.Map{
			"source": "database",
			"assets": assets,
		})
	}

	// Blockchain logic
	id, sign, err := fabric.GetIdentity("admin", h.WalletPath)
	if err != nil {
		id, sign, err = fabric.GetIdentity("test_user_post_tls", h.WalletPath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "No identity available"})
		}
	}

	// Type assert Conn
	grpcConn, ok := h.Conn.(*grpc.ClientConn)
	if !ok {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid gRPC connection type"})
	}

	gw, err := fabric.CreateGateway(grpcConn, id, sign)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer gw.Close()

	network := gw.GetNetwork(h.Config.ChannelName)
	contract := network.GetContract(h.Config.ChaincodeName)

	result, err := contract.EvaluateTransaction("GetAllAssets")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var assets []interface{}
	if err := json.Unmarshal(result, &assets); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse assets"})
	}

	return c.JSON(fiber.Map{
		"source": "blockchain",
		"assets": assets,
	})
}

func (h *AdminHandler) Sync(c *fiber.Ctx) error {
	// 1. Fetch all assets from Blockchain
	id, sign, err := fabric.GetIdentity("admin", h.WalletPath)
	if err != nil {
		id, sign, err = fabric.GetIdentity("test_user_post_tls", h.WalletPath)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "No identity available for sync"})
		}
	}

	grpcConn, ok := h.Conn.(*grpc.ClientConn)
	if !ok {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid gRPC connection"})
	}

	gw, err := fabric.CreateGateway(grpcConn, id, sign)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	defer gw.Close()

	network := gw.GetNetwork(h.Config.ChannelName)
	contract := network.GetContract(h.Config.ChaincodeName)

	result, err := contract.EvaluateTransaction("GetAllAssets")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var ledgerAssets []models.Asset
	if err := json.Unmarshal(result, &ledgerAssets); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse ledger assets"})
	}

	// 2. Sync to Database
	for _, asset := range ledgerAssets {
		if err := h.DB.Save(&asset).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to sync asset to DB: " + err.Error()})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Synchronization complete",
		"count":   len(ledgerAssets),
	})
}

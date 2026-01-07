package api

import (
	"backend/internal/fabric"
	"backend/internal/models"
	"encoding/json"
	"log"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type AdminHandler struct {
	CAConfigs  []fabric.CAConfig
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
	Status   string `json:"status"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Org      string `json:"org"`
}

func (h *AdminHandler) GetStats(c *fiber.Ctx) error {
	// ... (Existing GetStats)
	stats := NetworkStats{
		TotalAssets:      42,
		TotalOwners:      12,
		PendingTransfers: 5,
	}

	return c.JSON(stats)
}

func (h *AdminHandler) GetUsers(c *fiber.Ctx) error {
	var allIdentities []IdentityInfo

	// Fetch all DB users to match statuses
	var dbUsers []models.User
	h.DB.Find(&dbUsers)
	dbMap := make(map[string]models.User)
	for _, u := range dbUsers {
		dbMap[u.Username] = u
	}

	reName := regexp.MustCompile(`Name: ([^,]+)`)
	reType := regexp.MustCompile(`Type: ([^,]+)`)

	for _, caCfg := range h.CAConfigs {
		raw, err := fabric.ListIdentities(caCfg)
		if err != nil {
			log.Printf("Warning: failed to list identities from %s: %v", caCfg.MSPID, err)
			continue
		}

		lines := strings.Split(strings.TrimSpace(raw), "\n")
		for _, line := range lines {
			nameMatch := reName.FindStringSubmatch(line)
			typeMatch := reType.FindStringSubmatch(line)

			if len(nameMatch) > 1 && len(typeMatch) > 1 {
				name := nameMatch[1]
				status := "ACTIVE"
				email := name + "@example.org"
				role := "user"

				if dbUser, exists := dbMap[name]; exists {
					status = dbUser.Status
					email = dbUser.Email
					role = dbUser.Role
				}

				allIdentities = append(allIdentities, IdentityInfo{
					Name:     name,
					Type:     typeMatch[1],
					DBStatus: "Synced",
					Status:   status,
					Email:    email,
					Role:     role,
					Org:      caCfg.MSPID,
				})
			}
		}
	}

	return c.JSON(fiber.Map{"identities": allIdentities})
}

func (h *AdminHandler) UpdateUserStatus(c *fiber.Ctx) error {
	username := c.Params("username")
	type StatusUpdate struct {
		Status string `json:"status"` // ACTIVE, BANNED
		Role   string `json:"role"`   // user, auditor, admin
	}

	var req StatusUpdate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if username == "admin" {
		return c.Status(400).JSON(fiber.Map{"error": "Cannot modify root admin"})
	}

	// Update or Create DB record
	var user models.User
	err := h.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		user = models.User{
			Username: username,
			Email:    username + "@example.org",
			Status:   "ACTIVE",
			Role:     "user",
		}
	}

	if req.Status != "" {
		user.Status = req.Status
	}
	if req.Role != "" {
		user.Role = req.Role
	}

	if err := h.DB.Save(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update user status"})
	}

	return c.JSON(fiber.Map{"message": "User updated successfully", "status": user.Status, "role": user.Role})
}

func (h *AdminHandler) UpdateAssetStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	type StatusReq struct {
		Status string `json:"status"`
	}
	var req StatusReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// 1. Update on Blockchain
	identity, sign, err := fabric.GetIdentity("admin", h.WalletPath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Admin identity required for asset management"})
	}

	grpcConn, ok := h.Conn.(*grpc.ClientConn)
	if !ok {
		return c.Status(500).JSON(fiber.Map{"error": "Invalid connection"})
	}

	gw, err := fabric.CreateGateway(grpcConn, identity, sign)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Gateway creation failed"})
	}
	defer gw.Close()

	network := gw.GetNetwork(h.Config.ChannelName)
	contract := network.GetContract(h.Config.ChaincodeName)

	_, err = contract.SubmitTransaction("UpdateAssetStatus", id, req.Status)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Blockchain update failed: " + err.Error()})
	}

	// 2. Update in DB (if exists)
	var asset models.Asset
	if err := h.DB.Where("id = ?", id).First(&asset).Error; err == nil {
		asset.Status = req.Status
		h.DB.Save(&asset)
	}

	return c.JSON(fiber.Map{"message": "Asset status updated: " + req.Status, "status": req.Status})
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

	// Blockchain logic: Admin only
	id, sign, err := fabric.GetIdentity("admin", h.WalletPath)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "Admin identity not found in wallet"})
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
		return c.Status(401).JSON(fiber.Map{"error": "Admin identity not found in wallet. Please register 'admin' first."})
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

package api

import (
	"backend/internal/auth"
	"backend/internal/fabric"
	"backend/internal/models"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// AuthHandler holds dependencies
type AuthHandler struct {
	CAConfigs map[string]fabric.CAConfig
	DB        *gorm.DB
}

// Login handles user authentication (Enrollment + JWT)
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	type LoginReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Org      string `json:"org"` // Optional: If provided, only try this CA
	}

	var req LoginReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// 1. Find the correct CA and Enroll
	// If the user doesn't specify an org, we try all of them
	// In a real app, you'd probably have an org dropdown on login or derive it from the username
	token := ""
	var err error
	var finalCfg fabric.CAConfig

	for mspid, cfg := range h.CAConfigs {
		// If org provided, skip others
		if req.Org != "" && req.Org != mspid {
			continue
		}

		err = fabric.EnrollUser(cfg, req.Username, req.Password)
		if err == nil {
			finalCfg = cfg
			break
		}
	}

	if err != nil {
		log.Printf("Enroll failed for %s: %v", req.Username, err)
		return c.Status(401).JSON(fiber.Map{"error": "Authentication failed"})
	}

	// 2. Check DB Status
	var user models.User
	err = h.DB.Where("username = ? AND org = ?", req.Username, finalCfg.MSPID).First(&user).Error
	if err == nil {
		if user.Status == "BANNED" {
			return c.Status(403).JSON(fiber.Map{"error": "Your account has been banned. Contact Admin."})
		}
		if user.Status == "DELETED" {
			return c.Status(403).JSON(fiber.Map{"error": "This account has been deleted."})
		}
		// Backfill Org if missing
		if user.Org == "" && req.Username != "admin" {
			user.Org = finalCfg.MSPID
			h.DB.Save(&user)
		}
	} else {
		// If user exists in CA but not in DB, create it (backfill)
		userRole := "user"
		if req.Username == "admin" || strings.HasPrefix(req.Username, "admin-") {
			userRole = "admin"
		}

		// Generate email based on org to prevent collisions
		userEmail := req.Username + "@example.org"
		if finalCfg.MSPID == "Org2MSP" {
			userEmail = req.Username + "@org2.example.com"
		}

		user = models.User{
			Username: req.Username,
			Org:      finalCfg.MSPID,
			Email:    userEmail,
			Role:     userRole,
			Status:   "ACTIVE",
		}
		if err := h.DB.Create(&user).Error; err != nil {
			log.Printf("Failed to backfill DB profile for %s (%s): %v", req.Username, finalCfg.MSPID, err)
		}
	}

	// 3. Determine Role
	role := user.Role
	if req.Username == "admin" || strings.HasPrefix(req.Username, "admin-") {
		role = "admin" // Override safety for admin accounts
	}

	// 4. Finalize Response Org
	orgResp := user.Org
	if req.Username == "admin" {
		orgResp = "Org1MSP" // Default admin org
		if finalCfg.MSPID != "" {
			orgResp = finalCfg.MSPID
		}
	}

	// 5. Generate Dual Tokens
	token, err = auth.GenerateToken(req.Username, role, orgResp)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Access token generation failed"})
	}

	refreshToken, err := auth.GenerateRefreshToken(req.Username, role, orgResp)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Refresh token generation failed"})
	}

	// 6. Set HttpOnly Cookies
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    token,
		Expires:  time.Now().Add(15 * time.Minute),
		HTTPOnly: true,
		Secure:   false, // Set to true in production
		SameSite: "Lax",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HTTPOnly: true,
		Secure:   false, // Set to true in production
		SameSite: "Lax",
	})

	return c.JSON(fiber.Map{
		"username": req.Username,
		"org":      orgResp,
		"role":     role,
		"status":   user.Status,
		"token":    token, // Return for backward compatibility/header backup
	})
}

// Refresh handles token rotation using the long-lived refresh token
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(401).JSON(fiber.Map{"error": "Refresh token required"})
	}

	// Validate refresh token
	token, err := jwt.ParseWithClaims(refreshToken, &auth.UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("super-secret-key-change-in-production"), nil // Shared secret
	})

	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid or expired refresh token"})
	}

	claims, ok := token.Claims.(*auth.UserClaims)
	if !ok {
		return c.Status(401).JSON(fiber.Map{"error": "Invalid claims"})
	}

	// Generate new access token
	newAccessToken, err := auth.GenerateToken(claims.Username, claims.Role, claims.Org)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Token generation failed"})
	}

	// Update access token cookie
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    newAccessToken,
		Expires:  time.Now().Add(15 * time.Minute),
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})

	return c.JSON(fiber.Map{
		"token": newAccessToken,
	})
}

// Logout clears the authentication cookies
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	})
	return c.JSON(fiber.Map{"message": "Logged out successfully"})
}

// Register handles new user registration
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	type RegisterReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
		Org      string `json:"org"` // Org1MSP or Org2MSP
	}

	var req RegisterReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// TODO: get MSPID in smart contract
	// Determine CA Config
	org := req.Org
	if org == "" {
		org = "Org1MSP" // Default
	}

	cfg, ok := h.CAConfigs[org]
	if !ok {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid organization"})
	}

	// Call Fabric CA Registration
	resp, err := fabric.RegisterUser(cfg, req.Username, req.Password)
	if err != nil {
		log.Printf("Register failed for %s on %s: %v", req.Username, org, err)
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Registration failed: %v", err)})
	}

	// Create user in Off-Chain DB
	userEmail := req.Email
	if userEmail == "" {
		if org == "Org2MSP" {
			userEmail = req.Username + "@org2.example.com"
		} else {
			userEmail = req.Username + "@example.org"
		}
	}

	dbUser := models.User{
		Username: req.Username,
		Email:    userEmail,
		Org:      org,
		Role:     "user",
		Status:   "PENDING", // Wait for Admin Approval
	}
	if err := h.DB.Create(&dbUser).Error; err != nil {
		log.Printf("Failed to create DB profile for %s: %v", req.Username, err)
		// Don't fail the whole request because CA reg succeeded
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "User registered successfully. Status: PENDING (Awaiting Admin Approval)",
		"details": resp,
	})
}

// DeleteAccount allows a user to soft-delete their own account
func (h *AuthHandler) DeleteAccount(c *fiber.Ctx) error {
	username := c.Locals("user").(string)
	org := c.Locals("org").(string)
	fullID := fmt.Sprintf("%s::%s", org, username)

	if username == "admin" {
		return c.Status(400).JSON(fiber.Map{"error": "Root administrator account cannot be deleted"})
	}

	// 1. Integrity Check: Does the user own any active assets?
	var assetCount int64
	h.DB.Model(&models.Asset{}).Where("owner_id = ? AND status != ?", fullID, "DELETED").Count(&assetCount)
	if assetCount > 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": fmt.Sprintf("Account cannot be deleted while you own %d active artifacts. Please transfer or delete your artifacts first.", assetCount),
		})
	}

	// 2. Soft Delete in Database
	if err := h.DB.Model(&models.User{}).Where("username = ? AND org = ?", username, org).Update("status", "DELETED").Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to deactivate account"})
	}

	return c.JSON(fiber.Map{"message": "Account successfully deactivated."})
}

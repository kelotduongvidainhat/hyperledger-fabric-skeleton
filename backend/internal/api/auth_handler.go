package api

import (
	"backend/internal/auth"
	"backend/internal/fabric"
	"backend/internal/models"
	"log"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// AuthHandler holds dependencies
type AuthHandler struct {
	CAConfig fabric.CAConfig
	DB       *gorm.DB
}

// Login handles user authentication (Enrollment + JWT)
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	type LoginReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req LoginReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// 1. Enroll with CA (verify credentials)
	// This ensures the user exists and password is correct.
	// It also refreshes the certs in the wallet.
	err := fabric.EnrollUser(h.CAConfig, req.Username, req.Password)
	if err != nil {
		log.Printf("Enroll failed for %s: %v", req.Username, err)
		return c.Status(401).JSON(fiber.Map{"error": "Authentication failed"})
	}

	// 2. Check DB Status
	var user models.User
	err = h.DB.Where("username = ?", req.Username).First(&user).Error
	if err == nil {
		if user.Status == "BANNED" {
			return c.Status(403).JSON(fiber.Map{"error": "Your account has been banned. Contact Admin."})
		}
	} else if req.Username != "admin" {
		// If user exists in CA but not in DB, create it (backfill)
		user = models.User{
			Username: req.Username,
			Role:     "user",
			Status:   "ACTIVE", // Auto-active for previously registered users
			Email:    req.Username + "@example.org",
		}
		h.DB.Create(&user)
	}

	// 3. Determine Role
	role := "user"
	if req.Username == "admin" {
		role = "admin"
	} else if err == nil {
		role = user.Role
	}

	// 4. Generate JWT
	token, err := auth.GenerateToken(req.Username, role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Token generation failed"})
	}

	return c.JSON(fiber.Map{
		"token":    token,
		"username": req.Username,
		"role":     role,
		"status":   user.Status,
	})
}

// Register handles new user registration
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	type RegisterReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	var req RegisterReq
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Call Fabric CA Registration
	resp, err := fabric.RegisterUser(h.CAConfig, req.Username, req.Password)
	if err != nil {
		log.Printf("Register failed for %s: %v", req.Username, err)
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Registration failed: %v", err)})
	}

	// Create user in Off-Chain DB
	dbUser := models.User{
		Username: req.Username,
		Email:    req.Email,
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

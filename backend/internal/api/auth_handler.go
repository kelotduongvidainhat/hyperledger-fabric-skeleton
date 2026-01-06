package api

import (
	"backend/internal/auth"
	"backend/internal/fabric"
	"log"
	"fmt"
	"github.com/gofiber/fiber/v2"
)

// AuthHandler holds dependencies
type AuthHandler struct {
	CAConfig fabric.CAConfig
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

	// 2. Determine Role
	role := "user"
	if req.Username == "admin" {
		role = "admin"
	}

	// 3. Generate JWT
	token, err := auth.GenerateToken(req.Username, role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Token generation failed"})
	}

	return c.JSON(fiber.Map{
		"token":    token,
		"username": req.Username,
		"role":     role,
	})
}

// Register handles new user registration
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	type RegisterReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
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

	return c.Status(201).JSON(fiber.Map{
		"message": "User registered successfully",
		"details": resp,
	})
}

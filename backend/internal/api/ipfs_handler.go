package api

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	shell "github.com/ipfs/go-ipfs-api"
)

// IPFSHandler holds the IPFS shell dependency
type IPFSHandler struct {
	Shell *shell.Shell
}

// Upload handles file uploads to IPFS
func (h *IPFSHandler) Upload(c *fiber.Ctx) error {
	// Parse the multipart form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "No file uploaded"})
	}

	// Open the file
	file, err := fileHeader.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to open file"})
	}
	defer file.Close()

	// Upload to IPFS
	cid, err := h.Shell.Add(file)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("IPFS upload failed: %v", err)})
	}

	return c.JSON(fiber.Map{
		"message": "File uploaded to IPFS",
		"cid":     cid,
		"url":     fmt.Sprintf("ipfs://%s", cid),
	})
}

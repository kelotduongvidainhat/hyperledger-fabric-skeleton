package api

import (
	"backend/internal/models"
	"backend/internal/storage"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/gofiber/fiber/v2"
	shell "github.com/ipfs/go-ipfs-api"
	"strings"
)

type StorageHandler struct {
	MinIO *storage.MinIOStorage
	Ipfs  *shell.Shell
}

func (h *StorageHandler) Upload(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "No file uploaded"})
	}

	src, err := file.Open()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to open file"})
	}
	defer src.Close()

	// Read content to calculate hash and upload to multiple sinks
	content, err := io.ReadAll(src)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to read file"})
	}

	// 1. Calculate Hash
	hash := sha256.Sum256(content)
	fileHash := hex.EncodeToString(hash[:])

	// 2. Upload to IPFS
	var ipfsCID string
	if h.Ipfs != nil {
		ipfsCID, err = h.Ipfs.Add(bytes.NewReader(content))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("IPFS upload failed: %v", err)})
		}
	}

	// 3. Upload to MinIO
	var storagePath string
	// Sanitize filename: replace spaces with underscores to avoid URL encoding headaches
	safeFilename := strings.ReplaceAll(file.Filename, " ", "_")
	objectName := fmt.Sprintf("%d_%s", time.Now().Unix(), safeFilename)
	if h.MinIO != nil {
		storagePath, err = h.MinIO.Upload(objectName, bytes.NewReader(content), file.Size, file.Header.Get("Content-Type"))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("MinIO upload failed: %v", err)})
		}
	}

	// 4. Return the Attachment Metadata
	attachment := models.AssetAttachment{
		FileName:    file.Filename,
		FileSize:    file.Size,
		FileHash:    fileHash,
		IpfsCID:     ipfsCID,
		StoragePath: storagePath,
		StorageType: "minio",
	}

	return c.JSON(attachment)
}

func (h *StorageHandler) GetURL(c *fiber.Ctx) error {
	if h.MinIO == nil {
		return c.Status(503).JSON(fiber.Map{"error": "Storage service unavailable"})
	}

	objectName := c.Params("objectName")
	if objectName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Object name required"})
	}

	download := c.Query("download") == "true"
	url, err := h.MinIO.GetPresignedURL(objectName, 24*time.Hour, download)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": fmt.Sprintf("Failed to generate URL: %v", err)})
	}

	return c.JSON(fiber.Map{
		"url": url,
	})
}

package fabric

import (
	"backend/internal/models"
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/hyperledger/fabric-gateway/pkg/client"
	"gorm.io/gorm"
)

// StartEventListener connects to the blockchain and listens for events to sync the database
func StartEventListener(ctx context.Context, network *client.Network, db *gorm.DB) {
	log.Println("Starting Eventual Consistency Listener...")

	blocks, err := network.BlockEvents(ctx)
	if err != nil {
		log.Printf("Failed to subscribe to block events: %v", err)
		return
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping event listener...")
			return
		case block := <-blocks:
			if block == nil {
				continue
			}
			log.Printf("Received block %d. Processing for consistency...", block.Header.Number)
			
			syncLedgerToDB(network, db)
		}
	}
}

// syncLedgerToDB is a helper to pull all assets and update the DB
func syncLedgerToDB(network *client.Network, db *gorm.DB) {
	contract := network.GetContract("basic")
	
	result, err := contract.EvaluateTransaction("GetAllAssets")
	if err != nil {
		log.Printf("Sync Error: Failed to evaluate GetAllAssets: %v", err)
		return
	}

	var ledgerValues []models.LedgerValue
	if err := json.Unmarshal(result, &ledgerValues); err != nil {
		log.Printf("Sync Error: Failed to parse ledger values: %v", err)
		return
	}

	// Sync to DB
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

		if err := db.Save(&asset).Error; err != nil {
			log.Printf("Sync Error: Failed to save asset %s: %v", asset.ID, err)
		}
	}
	
	log.Printf("Eventual Consistency: Synced %d assets from ledger to database.", len(ledgerValues))
}

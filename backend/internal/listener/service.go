package listener

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

type Asset struct {
	ID             string `json:"ID"`
	Color          string `json:"Color"`
	Size           int    `json:"Size"`
	Owner          string `json:"Owner"`
	AppraisedValue int    `json:"AppraisedValue"`
}

// StartEventListener listens for chaincode events and updates the database
func StartEventListener(ctx context.Context, network *client.Network, db *sql.DB) error {
	log.Println("Starting Chaincode Event Listener...")

	events, err := network.ChaincodeEvents(ctx, "asset-transfer")
	if err != nil {
		return fmt.Errorf("failed to start chaincode event listener: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-events:
			if event == nil {
				continue
			}
			log.Printf("Received event: %s payload: %s\n", event.EventName, string(event.Payload))

			var asset Asset
			if err := json.Unmarshal(event.Payload, &asset); err != nil {
				log.Printf("Error unmarshaling event payload: %v\n", err)
				continue
			}

			if event.EventName == "AssetCreated" || event.EventName == "AssetTransferred" {
				// Upsert asset
				query := `
					INSERT INTO assets (id, color, size, owner, appraised_value, last_updated)
					VALUES ($1, $2, $3, $4, $5, NOW())
					ON CONFLICT (id) DO UPDATE SET
						color = EXCLUDED.color,
						size = EXCLUDED.size,
						owner = EXCLUDED.owner,
						appraised_value = EXCLUDED.appraised_value,
						last_updated = NOW();
				`
				_, err := db.Exec(query, asset.ID, asset.Color, asset.Size, asset.Owner, asset.AppraisedValue)
				if err != nil {
					log.Printf("Error updating database: %v\n", err)
				} else {
					log.Printf("Successfully synced asset %s to database\n", asset.ID)
				}
			}
		}
	}
}

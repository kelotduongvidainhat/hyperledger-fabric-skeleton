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
	ID        string `json:"ID"`
	Name      string `json:"Name"`
	Category  string `json:"Category"`
	Owner     string `json:"Owner"`
	Status    string `json:"Status"`
	Updated   string `json:"Updated"`
	UpdatedBy string `json:"UpdatedBy"`
}

type User struct {
	ID        string `json:"ID"`
	Name      string `json:"Name"`
	Role      string `json:"Role"`
	Status    string `json:"Status"`
	Updated   string `json:"Updated"`
	UpdatedBy string `json:"UpdatedBy"`
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

			if event.EventName == "AssetCreated" || event.EventName == "AssetTransferred" {
				var asset Asset
				if err := json.Unmarshal(event.Payload, &asset); err != nil {
					log.Printf("Error unmarshaling asset event payload: %v\n", err)
					continue
				}

				// Upsert asset
				query := `
					INSERT INTO assets (id, name, category, owner, status, updated, updated_by, last_updated)
					VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
					ON CONFLICT (id) DO UPDATE SET
						name = EXCLUDED.name,
						category = EXCLUDED.category,
						owner = EXCLUDED.owner,
						status = EXCLUDED.status,
						updated = EXCLUDED.updated,
						updated_by = EXCLUDED.updated_by,
						last_updated = NOW();
				`
				_, err := db.Exec(query, asset.ID, asset.Name, asset.Category, asset.Owner, asset.Status, asset.Updated, asset.UpdatedBy)
				if err != nil {
					log.Printf("Error updating database (asset): %v\n", err)
				} else {
					log.Printf("Successfully synced asset %s to database\n", asset.ID)
				}
			} else if event.EventName == "UserCreated" {
				var user User
				if err := json.Unmarshal(event.Payload, &user); err != nil {
					log.Printf("Error unmarshaling user event payload: %v\n", err)
					continue
				}

				// Upsert user
				query := `
					INSERT INTO users (id, name, role, status, updated, updated_by, last_updated)
					VALUES ($1, $2, $3, $4, $5, $6, NOW())
					ON CONFLICT (id) DO UPDATE SET
						name = EXCLUDED.name,
						role = EXCLUDED.role,
						status = EXCLUDED.status,
						updated = EXCLUDED.updated,
						updated_by = EXCLUDED.updated_by,
						last_updated = NOW();
				`
				_, err := db.Exec(query, user.ID, user.Name, user.Role, user.Status, user.Updated, user.UpdatedBy)
				if err != nil {
					log.Printf("Error updating database (user): %v\n", err)
				} else {
					log.Printf("Successfully synced user %s to database\n", user.ID)
				}
			}
		}
	}
}

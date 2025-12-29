package fabric

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-gateway/pkg/client"
)

// Asset describes basic details of an asset
type Asset struct {
	ID             string `json:"ID"`
	Color          string `json:"Color"`
	Size           int    `json:"Size"`
	Owner          string `json:"Owner"`
	AppraisedValue int    `json:"AppraisedValue"`
}

// GetAllAssets returns all assets from the ledger
func (f *FabricClient) GetAllAssets(userID string) ([]Asset, error) {
	fmt.Printf("Evaluating GetAllAssets as %s...\n", userID)
	result, err := f.executeAction(userID, func(c *client.Contract) (interface{}, error) {
		return c.EvaluateTransaction("GetAllAssets")
	})
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate transaction: %w", err)
	}

	var assets []Asset
	err = json.Unmarshal(result.([]byte), &assets)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return assets, nil
}

// CreateAsset creates a new asset on the ledger
func (f *FabricClient) CreateAsset(userID, id, color string, size int, owner string, value int) error {
	fmt.Printf("Submitting CreateAsset: %s as %s...\n", id, userID)
	_, err := f.executeAction(userID, func(c *client.Contract) (interface{}, error) {
		return c.SubmitTransaction("CreateAsset", id, color, fmt.Sprint(size), owner, fmt.Sprint(value))
	})
	if err != nil {
		return fmt.Errorf("failed to submit transaction: %w", err)
	}

	return nil
}

// ReadAsset returns the asset with the specified ID
func (f *FabricClient) ReadAsset(userID, id string) (*Asset, error) {
	fmt.Printf("Evaluating ReadAsset: %s as %s...\n", id, userID)
	result, err := f.executeAction(userID, func(c *client.Contract) (interface{}, error) {
		return c.EvaluateTransaction("ReadAsset", id)
	})
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate transaction: %w", err)
	}

	var asset Asset
	err = json.Unmarshal(result.([]byte), &asset)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &asset, nil
}

// TransferAsset updates the owner of an asset
func (f *FabricClient) TransferAsset(userID, id string, newOwner string) error {
	fmt.Printf("Submitting TransferAsset: %s to %s as %s...\n", id, newOwner, userID)
	_, err := f.executeAction(userID, func(c *client.Contract) (interface{}, error) {
		return c.SubmitTransaction("TransferAsset", id, newOwner)
	})
	if err != nil {
		return fmt.Errorf("failed to submit transaction: %w", err)
	}

	return nil
}

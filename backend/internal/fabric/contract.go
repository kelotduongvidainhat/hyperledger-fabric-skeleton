package fabric

import (
	"encoding/json"
	"fmt"
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
func (f *FabricClient) GetAllAssets() ([]Asset, error) {
	fmt.Println("Evaluating GetAllAssets...")
	evaluateResponse, err := f.Contract.EvaluateTransaction("GetAllAssets")
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate transaction: %w", err)
	}

	var assets []Asset
	err = json.Unmarshal(evaluateResponse, &assets)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return assets, nil
}

// CreateAsset creates a new asset on the ledger
func (f *FabricClient) CreateAsset(id, color string, size int, owner string, value int) error {
	fmt.Printf("Submitting CreateAsset: %s...\n", id)
	_, err := f.Contract.SubmitTransaction("CreateAsset", id, color, fmt.Sprint(size), owner, fmt.Sprint(value))
	if err != nil {
		return fmt.Errorf("failed to submit transaction: %w", err)
	}

	return nil
}

// ReadAsset returns the asset with the specified ID
func (f *FabricClient) ReadAsset(id string) (*Asset, error) {
	fmt.Printf("Evaluating ReadAsset: %s...\n", id)
	evaluateResponse, err := f.Contract.EvaluateTransaction("ReadAsset", id)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate transaction: %w", err)
	}

	var asset Asset
	err = json.Unmarshal(evaluateResponse, &asset)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &asset, nil
}

// TransferAsset updates the owner of an asset
func (f *FabricClient) TransferAsset(id string, newOwner string) error {
	fmt.Printf("Submitting TransferAsset: %s to %s...\n", id, newOwner)
	_, err := f.Contract.SubmitTransaction("TransferAsset", id, newOwner)
	if err != nil {
		return fmt.Errorf("failed to submit transaction: %w", err)
	}

	return nil
}

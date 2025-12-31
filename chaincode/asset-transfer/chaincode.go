package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
// Asset describes basic details of what makes up a simple asset
type Asset struct {
	ID             string `json:"ID"`
	Name           string `json:"Name"`
	Category       string `json:"Category"`
	Owner          string `json:"Owner"`		// User ID
	Status         string `json:"Status"` 		// AVAILABLE, FROZEN, DELETED
	Updated        string `json:"Updated"` 		// Date Time YYYY-MM-DD HH:MM:SS
	UpdatedBy      string `json:"UpdatedBy"`	// User ID
}

type User struct {
	ID             string `json:"ID"`
	Name           string `json:"Name"`
	Role           string `json:"Role"`			// ADMIN, USER
	Status         string `json:"Status"`		// AVAILABLE, FROZEN, DELETED
	Updated        string `json:"Updated"`		// Date Time YYYY-MM-DD HH:MM:SS
	UpdatedBy      string `json:"UpdatedBy"`	// User ID
}

const (
	AssetPrefix = "asset_"
	UserPrefix  = "user_"
)

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{ID: "asset1", Name: "asset1", Category: "Art", Owner: "Tomoko", Status: "AVAILABLE", Updated: "2025-12-31 12:00:00", UpdatedBy: "Tomoko"},
		{ID: "asset2", Name: "asset2", Category: "Car", Owner: "Brad", Status: "AVAILABLE", Updated: "2025-12-31 12:00:00", UpdatedBy: "Brad"},
		{ID: "asset3", Name: "asset3", Category: "Art", Owner: "Jin Soo", Status: "AVAILABLE", Updated: "2025-12-31 12:00:00", UpdatedBy: "Jin Soo"},
		{ID: "asset4", Name: "asset4", Category: "Car", Owner: "Max", Status: "AVAILABLE", Updated: "2025-12-31 12:00:00", UpdatedBy: "Max"},
		{ID: "asset5", Name: "asset5", Category: "Apartment", Owner: "Adriana", Status: "AVAILABLE", Updated: "2025-12-31 12:00:00", UpdatedBy: "Adriana"},
		{ID: "asset6", Name: "asset6", Category: "Phone", Owner: "Michel", Status: "AVAILABLE", Updated: "2025-12-31 12:00:00", UpdatedBy: "Michel"},
	}
	users := []User{
		{ID: "user1", Name: "user1", Role: "ADMIN", Status: "AVAILABLE", Updated: "2025-12-31 12:00:00", UpdatedBy: "user1"},
		{ID: "user2", Name: "user2", Role: "USER", Status: "AVAILABLE", Updated: "2025-12-31 12:00:00", UpdatedBy: "user2"},
		{ID: "user3", Name: "user3", Role: "USER", Status: "AVAILABLE", Updated: "2025-12-31 12:00:00", UpdatedBy: "user3"},
		{ID: "user4", Name: "user4", Role: "USER", Status: "AVAILABLE", Updated: "2025-12-31 12:00:00", UpdatedBy: "user4"},
		{ID: "user5", Name: "user5", Role: "USER", Status: "AVAILABLE", Updated: "2025-12-31 12:00:00", UpdatedBy: "user5"},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(AssetPrefix+asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state: %v", err)
		} 
		ctx.GetStub().SetEvent("AssetCreated", assetJSON)
		
	}

	for _, user := range users {
		userJSON, err := json.Marshal(user)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(UserPrefix+user.ID, userJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state: %v", err)
		} 
		ctx.GetStub().SetEvent("UserCreated", userJSON)
		
	}

	return nil
}

// CreateAsset issues a new asset to the world state with given details
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, name string, category string, owner string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	timestamp, err := s.getTxTimestamp(ctx)
	if err != nil {
		return err
	}

	creator, err := s.getCallerID(ctx)
	if err != nil {
		return err
	}

	asset := Asset{
		ID:             id,
		Name:           name,
		Category:       category,
		Owner:          owner,
		Status:         "AVAILABLE",
		Updated:        timestamp,
		UpdatedBy:      creator,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	// Emit event for backend listener
	err = ctx.GetStub().SetEvent("AssetCreated", assetJSON)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(AssetPrefix+id, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given id
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(AssetPrefix + id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, name string, category string, owner string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	timestamp, err := s.getTxTimestamp(ctx)
	if err != nil {
		return err
	}

	creator, err := s.getCallerID(ctx)
	if err != nil {
		return err
	}

	asset.Name = name
	asset.Category = category
	asset.Owner = owner
	asset.Status = "AVAILABLE"
	asset.Updated = timestamp
	asset.UpdatedBy = creator
	
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(AssetPrefix+id, assetJSON)
}

// DeleteAsset deletes an given asset from the world state
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(AssetPrefix + id)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(AssetPrefix + id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// TransferAsset updates the owner field of asset with given id in world state
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	asset.Owner = newOwner
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	// Emit event for backend listener
	err = ctx.GetStub().SetEvent("AssetTransferred", assetJSON)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(AssetPrefix+id, assetJSON)
}

// LockAsset sets the asset status to FROZEN
func (s *SmartContract) LockAsset(ctx contractapi.TransactionContextInterface, id string) error {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	asset.Status = "FROZEN"
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(AssetPrefix+id, assetJSON)
}

// UnlockAsset sets the asset status to AVAILABLE
func (s *SmartContract) UnlockAsset(ctx contractapi.TransactionContextInterface, id string) error {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	asset.Status = "AVAILABLE"
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(AssetPrefix+id, assetJSON)
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// Range query with empty string returns everything. We need to filter manually or change approach.
	// Implementing range query on AssetPrefix would be ideal if supported, but GetStateByRange works alphabetically.
	// Queries for "asset_" to "asset_~" (~ is high ASCII)
	resultsIterator, err := ctx.GetStub().GetStateByRange(AssetPrefix, AssetPrefix+"~")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

// ----------------- USER FUNCTIONS -----------------

// CreateUser registers a new user
func (s *SmartContract) CreateUser(ctx contractapi.TransactionContextInterface, id string, name string, role string) error {
	exists, err := s.UserExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the user %s already exists", id)
	}

	timestamp, err := s.getTxTimestamp(ctx)
	if err != nil {
		return err
	}

	creator, err := s.getCallerID(ctx)
	if err != nil {
		return err
	}

	user := User{
		ID:        id,
		Name:      name,
		Role:      role,
		Status:    "AVAILABLE",
		Updated:   timestamp,
		UpdatedBy: creator,
	}
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	ctx.GetStub().SetEvent("UserCreated", userJSON)
	return ctx.GetStub().PutState(UserPrefix+id, userJSON)
}

// ReadUser returns the user stored in the world state with given id
func (s *SmartContract) ReadUser(ctx contractapi.TransactionContextInterface, id string) (*User, error) {
	userJSON, err := ctx.GetStub().GetState(UserPrefix + id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if userJSON == nil {
		return nil, fmt.Errorf("the user %s does not exist", id)
	}

	var user User
	err = json.Unmarshal(userJSON, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser updates an existing user
func (s *SmartContract) UpdateUser(ctx contractapi.TransactionContextInterface, id string, name string, role string) error {
	user, err := s.ReadUser(ctx, id)
	if err != nil {
		return err
	}

	timestamp, err := s.getTxTimestamp(ctx)
	if err != nil {
		return err
	}

	creator, err := s.getCallerID(ctx)
	if err != nil {
		return err
	}

	user.Name = name
	user.Role = role
	user.Updated = timestamp
	user.UpdatedBy = creator

	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(UserPrefix+id, userJSON)
}

// DeleteUser deletes a user from the world state
func (s *SmartContract) DeleteUser(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.UserExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the user %s does not exist", id)
	}

	return ctx.GetStub().DelState(UserPrefix + id)
}

// UserExists returns true when user with given ID exists in world state
func (s *SmartContract) UserExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	userJSON, err := ctx.GetStub().GetState(UserPrefix + id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return userJSON != nil, nil
}

// GetAllUsers returns all users found in world state
func (s *SmartContract) GetAllUsers(ctx contractapi.TransactionContextInterface) ([]*User, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange(UserPrefix, UserPrefix+"~")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var users []*User
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var user User
		err = json.Unmarshal(queryResponse.Value, &user)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

// LockUser sets the user status to FROZEN
func (s *SmartContract) LockUser(ctx contractapi.TransactionContextInterface, id string) error {
	user, err := s.ReadUser(ctx, id)
	if err != nil {
		return err
	}

	user.Status = "FROZEN"
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(UserPrefix+id, userJSON)
}

// UnlockUser sets the user status to AVAILABLE
func (s *SmartContract) UnlockUser(ctx contractapi.TransactionContextInterface, id string) error {
	user, err := s.ReadUser(ctx, id)
	if err != nil {
		return err
	}

	user.Status = "AVAILABLE"
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(UserPrefix+id, userJSON)
}

// HistoryRecord structure used for returning the history of an asset
type HistoryRecord struct {
	TxId      string `json:"txId"`
	Timestamp string `json:"timestamp"`
	IsDelete  bool   `json:"isDelete"`
	Record    *Asset `json:"record"`
}

// GetAssetHistory returns the chain of custody for an asset since issuance.
func (s *SmartContract) GetAssetHistory(ctx contractapi.TransactionContextInterface, assetID string) ([]HistoryRecord, error) {
	log.Printf("GetAssetHistory: ID %v", assetID)

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(AssetPrefix + assetID)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryRecord
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &asset)
			if err != nil {
				return nil, err
			}
		}

		timestamp, err := response.Timestamp.AsTime(), error(nil)
		if err != nil {
			return nil, err
		}

		record := HistoryRecord{
			TxId:      response.TxId,
			Timestamp: timestamp.String(),
			IsDelete:  response.IsDelete,
			Record:    &asset,
		}
		records = append(records, record)
	}

	return records, nil
}

// getTxTimestamp retrieves the transaction timestamp from the context and formats it
func (s *SmartContract) getTxTimestamp(ctx contractapi.TransactionContextInterface) (string, error) {
	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return "", err
	}
	// Convert protobuf timestamp to Go time.Time
	// Seconds is int64, Nanos is int32
	tm := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos))
	return tm.Format("2006-01-02 15:04:05"), nil
}

// getCallerID extracts the Common Name (CN) from the client's X509 certificate
func (s *SmartContract) getCallerID(ctx contractapi.TransactionContextInterface) (string, error) {
	// Get the client identity
	clientIdentity := ctx.GetClientIdentity()
	if clientIdentity == nil {
		return "", fmt.Errorf("failed to get client identity")
	}

	// Get the X.509 certificate
	cert, err := clientIdentity.GetX509Certificate()
	if err != nil {
		return "", fmt.Errorf("failed to get x509 certificate: %v", err)
	}

	// Return the Common Name (e.g., "user1", "admin")
	return cert.Subject.CommonName, nil
}

func main() {
	// Create the chaincode
	assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating asset-transfer chaincode: %v", err)
	}

	// Start as external chaincode service
	server := &shim.ChaincodeServer{
		CCID:    os.Getenv("CHAINCODE_ID"),
		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
		CC:      assetChaincode,
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	if err := server.Start(); err != nil {
		log.Panicf("Error starting asset-transfer chaincode: %v", err)
	}
}

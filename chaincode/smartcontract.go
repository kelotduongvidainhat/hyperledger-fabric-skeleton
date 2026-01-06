package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
// Insert struct field in alphabetic order to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't matter as much for this
type Asset struct {
	ID              string `json:"ID"`
	Name            string `json:"Name"`
	Description     string `json:"Description"`
	OwnerID         string `json:"OwnerID"` // Org1MSP::username
	ProposedOwnerID string `json:"ProposedOwnerID"` // Org1MSP::username
	ImageURL        string `json:"ImageURL"`
	ImageHash       string `json:"ImageHash"`
	Status          string `json:"Status"` // ACTIVE, FROZEN, DELETED, PENDING_TRANSFER
	View            string `json:"View"`
	LastUpdatedBy   string `json:"LastUpdatedBy"`
	LastUpdatedAt   string `json:"LastUpdatedAt"`
}

// HistoryRecord structure for returning asset history
type HistoryRecord struct {
	TxId       string    `json:"TxId"`
	Timestamp  time.Time `json:"Timestamp"`
	ActorID    string    `json:"ActorID"`
	ActionType string    `json:"ActionType"` // Custom action type
	Value      *Asset    `json:"Value"`
	IsDelete   bool      `json:"IsDelete"`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	txTimestamp, _ := ctx.GetStub().GetTxTimestamp()
	now := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).Format(time.RFC3339)

	assets := []Asset{
		{ID: "asset1", Name: "Genesis Asset", Description: "First Asset", OwnerID: "Org1MSP::admin", Status: "ACTIVE", View: "Public", LastUpdatedBy: "Init", LastUpdatedAt: now},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

func (s *SmartContract) getClientFullIdentifier(ctx contractapi.TransactionContextInterface) (string, error) {
	// 1. Lấy MSPID
	mspid, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to get MSPID: %v", err)
	}

	// 2. Thử lấy EnrollmentID (Dùng cho Fabric CA)
	username, found, _ := ctx.GetClientIdentity().GetAttributeValue("hf.EnrollmentID")
	
	// 3. Fallback: Nếu không thấy EnrollmentID, lấy CommonName (CN) từ Certificate Subject
	if !found || username == "" {
		cert, err := ctx.GetClientIdentity().GetX509Certificate()
		if err != nil {
			return "", fmt.Errorf("failed to get certificate: %v", err)
		}
		username = cert.Subject.CommonName
	}

	// 4. Kiểm tra cuối cùng nếu vẫn trống
	if username == "" {
		return mspid + "::unknown_identity", nil
	}

	return fmt.Sprintf("%s::%s", mspid, username), nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, name string, description string, imageURL string, imageHash string, view string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	clientFullID, err := s.getClientFullIdentifier(ctx)
	if err != nil {
		return err
	}

	txTimestamp, _ := ctx.GetStub().GetTxTimestamp()
	now := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).Format(time.RFC3339)

	asset := Asset{
		ID:              id,
		Name:            name,
		Description:     description,
		OwnerID:         clientFullID,
		ProposedOwnerID: "",
		ImageURL:        imageURL,
		ImageHash:       imageHash,
		Status:          "ACTIVE",
		View:            view,
		LastUpdatedBy:   clientFullID,
		LastUpdatedAt:   now,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("asset %s does not exist", id)
	}
	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal asset: %v", err)
	}
	return &asset, nil
}

// ProposeTransfer initiates the Two-Factor Transfer workflow.
func (s *SmartContract) ProposeTransfer(ctx contractapi.TransactionContextInterface, id string, newOwnerID string) error {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	clientFullID, err := s.getClientFullIdentifier(ctx)
	if err != nil {
		return err
	}

	if asset.OwnerID != clientFullID {
		return fmt.Errorf("only the owner can propose a transfer (current owner: %s, you: %s)", asset.OwnerID, clientFullID)
	}
	if asset.Status != "ACTIVE" {
		return fmt.Errorf("asset is not ACTIVE")
	}

	txTimestamp, _ := ctx.GetStub().GetTxTimestamp()
	now := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).Format(time.RFC3339)

	asset.Status = "PENDING_TRANSFER"
	asset.ProposedOwnerID = newOwnerID
	asset.LastUpdatedBy = clientFullID
	asset.LastUpdatedAt = now

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// AcceptTransfer finalizes the Two-Factor Transfer workflow.
func (s *SmartContract) AcceptTransfer(ctx contractapi.TransactionContextInterface, id string) error {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	clientFullID, err := s.getClientFullIdentifier(ctx)
	if err != nil {
		return err
	}

	if asset.Status != "PENDING_TRANSFER" {
		return fmt.Errorf("asset is not in PENDING_TRANSFER state")
	}
	if asset.ProposedOwnerID != clientFullID {
		return fmt.Errorf("you are not the proposed owner (expected: %s, you: %s)", asset.ProposedOwnerID, clientFullID)
	}

	txTimestamp, _ := ctx.GetStub().GetTxTimestamp()
	now := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).Format(time.RFC3339)

	asset.OwnerID = asset.ProposedOwnerID
	asset.ProposedOwnerID = ""
	asset.Status = "ACTIVE"
	asset.LastUpdatedBy = clientFullID
	asset.LastUpdatedAt = now

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// UpdateAssetStatus allows an authority to change the status of an asset (e.g., FROZEN, ACTIVE, DELETED)
func (s *SmartContract) UpdateAssetStatus(ctx contractapi.TransactionContextInterface, id string, newStatus string) error {
	// 1. SECURITY: Assert Admin status (using CID library)
	// In production, users should be registered with an 'admin' attribute
	// For this skeleton, we also allow the default Peer Admin (Org1MSP/Org2MSP)
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return err
	}

	// Dynamic check for admin attribute
	isAdmin, found, _ := ctx.GetClientIdentity().GetAttributeValue("admin")
	if (!found || isAdmin != "true") && clientMSPID != "Org1MSP" && clientMSPID != "Org2MSP" {
		return fmt.Errorf("administrative access required for this operation")
	}

	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	clientFullID, err := s.getClientFullIdentifier(ctx)
	if err != nil {
		return err
	}

	txTimestamp, _ := ctx.GetStub().GetTxTimestamp()
	now := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).Format(time.RFC3339)

	asset.Status = newStatus
	asset.LastUpdatedBy = clientFullID + "_ADMIN"
	asset.LastUpdatedAt = now

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// GetAssetHistory returns the provenance of an asset
func (s *SmartContract) GetAssetHistory(ctx contractapi.TransactionContextInterface, id string) ([]HistoryRecord, error) {
	historyIterator, err := ctx.GetStub().GetHistoryForKey(id)
	if err != nil {
		return nil, err
	}
	defer historyIterator.Close()

	var records []HistoryRecord
	for historyIterator.HasNext() {
		response, err := historyIterator.Next()
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

		// Determine Action Type based on status changes (heuristic)
		actionType := "UPDATE"
		if response.IsDelete {
			actionType = "DELETE"
		} else if asset.LastUpdatedBy == "Init" {
			actionType = "INIT"
		} else if asset.Status == "PENDING_TRANSFER" {
			actionType = "TRANSFER_PROPOSE"
		} else if asset.Status == "ACTIVE" && asset.LastUpdatedBy == asset.OwnerID {
			// This logic is simple; real world might store ActionType in the asset or use events
			actionType = "TRANSFER_ACCEPT_OR_CREATE" 
		}

		record := HistoryRecord{
			TxId:       response.TxId,
			Timestamp:  time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)),
			ActorID:    asset.LastUpdatedBy, // We use the stored field for Who
			ActionType: actionType,
			Value:      &asset,
			IsDelete:   response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
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

func main() {
	assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating ownership-registry chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting ownership-registry chaincode: %v", err)
	}
}

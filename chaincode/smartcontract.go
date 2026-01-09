package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	InitActorID = "Init"
	EmptyTxt    = ""
)

// Constants for asset status
const (
	ActiveStatus          = "ACTIVE"
	FrozenStatus          = "FROZEN"
	DeletedStatus         = "DELETED"
	PendingTransferStatus = "PENDING_TRANSFER"
)

// Constants for asset view
const (
	PublicView  = "PUBLIC"
	PrivateView = "PRIVATE"
)

// Constants for action type
const (
	InitActionType        = "INIT"
	CreateActionType      = "CREATE"
	UpdateStatusActionType = "UPDATE_STATUS"
	UpdateViewActionType  = "UPDATE_VIEW"
	DeleteActionType      = "DELETE"
	TransferProposeActionType = "TRANSFER_PROPOSE"
	TransferAcceptActionType  = "TRANSFER_ACCEPT"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

type AssetAttachment struct {
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	FileHash    string `json:"file_hash"` // SHA-256 of the file
	IpfsCID     string `json:"ipfs_cid"`
	StoragePath string `json:"storage_path"`
	StorageType string `json:"storage_type"` // Example: "minio", "s3", "local"
}

// Asset describes basic details of what makes up a simple asset (Lower-case tags for standard JSON)
type Asset struct {
	ID              string `json:"ID"` // Keep ID as ID for key consistency
	Name            string `json:"name"`
	Description     string `json:"description"`
	OwnerID         string `json:"ownerId"` 
	ProposedOwnerID string `json:"proposedOwnerId"`
	ImageURL        string `json:"imageUrl"`
	ImageHash       string `json:"imageHash"`
	Status          string `json:"status"` 
	View            string `json:"view"` 
	Attachment      AssetAttachment `json:"attachment"`
}

// AuditMetadata contains the metadata for a state change
type AuditMetadata struct {
	Action    string `json:"action"`
	Actor     string `json:"actor"`
	Timestamp string `json:"timestamp"`
}

// LedgerValue is the wrapper for data stored on the blockchain
type LedgerValue struct {
	Asset Asset         `json:"asset"`
	Audit AuditMetadata `json:"audit"`
}

// HistoryRecord structure for returning asset history
type HistoryRecord struct {
	TxId       string    `json:"txId"`
	Timestamp  time.Time `json:"timestamp"`
	ActorID    string    `json:"actorId"`
	ActionType string    `json:"actionType"`
	Value      *Asset    `json:"value"`
	IsDelete   bool      `json:"isDelete"`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	txTimestamp, _ := ctx.GetStub().GetTxTimestamp()
	now := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).Format(time.RFC3339)

	asset := Asset{ID: "asset1", Name: "Genesis Asset", Description: "First Asset", OwnerID: "Org1MSP::admin", Status: ActiveStatus, View: PublicView}
	
	ledgerValue := LedgerValue{
		Asset: asset,
		Audit: AuditMetadata{
			Action:    InitActionType,
			Actor:     InitActorID,
			Timestamp: now,
		},
	}

	valueJSON, err := json.Marshal(ledgerValue)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(asset.ID, valueJSON)
}

func (s *SmartContract) getClientFullIdentifier(ctx contractapi.TransactionContextInterface) (string, error) {
	mspid, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to get MSPID: %v", err)
	}

	username, found, _ := ctx.GetClientIdentity().GetAttributeValue("hf.EnrollmentID")
	if !found || username == EmptyTxt {
		cert, err := ctx.GetClientIdentity().GetX509Certificate()
		if err != nil {
			return EmptyTxt, fmt.Errorf("failed to get certificate: %v", err)
		}
		username = cert.Subject.CommonName
	}

	if username == EmptyTxt {
		return mspid + "::unknown_identity", nil
	}

	return fmt.Sprintf("%s::%s", mspid, username), nil
}

// CreateAsset issues a new asset to the world state
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, name string, description string, imageURL string, imageHash string, view string, 
	fileName string, fileSize int64, fileHash string, ipfsCID string, storagePath string, storageType string) error {
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
		ProposedOwnerID: EmptyTxt,
		ImageURL:        imageURL,
		ImageHash:       imageHash,
		Status:          ActiveStatus,
		View:            view,
		Attachment: AssetAttachment{
			FileName:    fileName,
			FileSize:    fileSize,
			FileHash:    fileHash,
			IpfsCID:     ipfsCID,
			StoragePath: storagePath,
			StorageType: storageType,
		},
	}

	ledgerValue := LedgerValue{
		Asset: asset,
		Audit: AuditMetadata{
			Action:    CreateActionType,
			Actor:     clientFullID,
			Timestamp: now,
		},
	}

	valueJSON, err := json.Marshal(ledgerValue)
	if err != nil {
		return err
	}

	err = ctx.GetStub().SetEvent("CreateAsset", valueJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	return ctx.GetStub().PutState(id, valueJSON)
}

func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*LedgerValue, error) {
	valueJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %v", err)
	}
	if valueJSON == nil {
		return nil, fmt.Errorf("asset %s does not exist", id)
	}
	var value LedgerValue
	
	var temp map[string]interface{}
	json.Unmarshal(valueJSON, &temp)
	
	// Flexible check for "asset" or "Asset"
	if _, ok := temp["asset"]; ok || temp["Asset"] != nil {
		err = json.Unmarshal(valueJSON, &value)
	} else {
		// Legacy fallback
		var asset Asset
		json.Unmarshal(valueJSON, &asset)
		value.Asset = asset
		value.Audit = AuditMetadata{Action: "LEGACY", Actor: "Unknown", Timestamp: ""}
	}
	
	return &value, nil
}

// ProposeTransfer initiates the Two-Factor Transfer workflow.
func (s *SmartContract) ProposeTransfer(ctx contractapi.TransactionContextInterface, id string, newOwnerID string) error {
	value, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	clientFullID, err := s.getClientFullIdentifier(ctx)
	if err != nil {
		return err
	}

	if value.Asset.OwnerID != clientFullID {
		return fmt.Errorf("only the owner can propose a transfer")
	}
	if value.Asset.Status != ActiveStatus {
		return fmt.Errorf("asset is not ACTIVE")
	}

	txTimestamp, _ := ctx.GetStub().GetTxTimestamp()
	now := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).Format(time.RFC3339)

	value.Asset.Status = PendingTransferStatus
	value.Asset.ProposedOwnerID = newOwnerID
	
	value.Audit = AuditMetadata{
		Action:    TransferProposeActionType,
		Actor:     clientFullID,
		Timestamp: now,
	}

	valueJSON, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = ctx.GetStub().SetEvent("ProposeTransfer", valueJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	return ctx.GetStub().PutState(id, valueJSON)
}

// AcceptTransfer finalizes the Two-Factor Transfer workflow.
func (s *SmartContract) AcceptTransfer(ctx contractapi.TransactionContextInterface, id string) error {
	value, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	clientFullID, err := s.getClientFullIdentifier(ctx)
	if err != nil {
		return err
	}

	if value.Asset.Status != PendingTransferStatus {
		return fmt.Errorf("asset is not in PENDING_TRANSFER state")
	}
	if value.Asset.ProposedOwnerID != clientFullID {
		return fmt.Errorf("you are not the proposed owner")
	}

	txTimestamp, _ := ctx.GetStub().GetTxTimestamp()
	now := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).Format(time.RFC3339)

	value.Asset.OwnerID = value.Asset.ProposedOwnerID
	value.Asset.ProposedOwnerID = ""
	value.Asset.Status = ActiveStatus
	
	value.Audit = AuditMetadata{
		Action:    TransferAcceptActionType,
		Actor:     clientFullID,
		Timestamp: now,
	}

	valueJSON, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = ctx.GetStub().SetEvent("AcceptTransfer", valueJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	return ctx.GetStub().PutState(id, valueJSON)
}

// UpdateAssetStatus allows an authority to change the status of an asset
func (s *SmartContract) UpdateAssetStatus(ctx contractapi.TransactionContextInterface, id string, newStatus string) error {
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return err
	}

	isAdmin, found, _ := ctx.GetClientIdentity().GetAttributeValue("admin")
	if (!found || isAdmin != "true") && clientMSPID != "Org1MSP" && clientMSPID != "Org2MSP" {
		return fmt.Errorf("administrative access required")
	}

	value, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	clientFullID, err := s.getClientFullIdentifier(ctx)
	if err != nil {
		return err
	}

	txTimestamp, _ := ctx.GetStub().GetTxTimestamp()
	now := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).Format(time.RFC3339)

	value.Asset.Status = newStatus
	value.Audit = AuditMetadata{
		Action:    UpdateStatusActionType,
		Actor:     clientFullID + "_ADMIN",
		Timestamp: now,
	}

	valueJSON, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = ctx.GetStub().SetEvent("UpdateAssetStatus", valueJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	return ctx.GetStub().PutState(id, valueJSON)
}

// UpdateAssetView allows the owner to change the visibility of an asset
func (s *SmartContract) UpdateAssetView(ctx contractapi.TransactionContextInterface, id string, newView string) error {
	value, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	clientFullID, err := s.getClientFullIdentifier(ctx)
	if err != nil {
		return err
	}

	if value.Asset.OwnerID != clientFullID {
		return fmt.Errorf("only the owner can change asset visibility")
	}

	if newView != PublicView && newView != PrivateView {
		return fmt.Errorf("invalid view status")
	}

	txTimestamp, _ := ctx.GetStub().GetTxTimestamp()
	now := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).Format(time.RFC3339)

	value.Asset.View = newView
	value.Audit = AuditMetadata{
		Action:    UpdateViewActionType,
		Actor:     clientFullID,
		Timestamp: now,
	}

	valueJSON, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = ctx.GetStub().SetEvent("UpdateAssetView", valueJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	return ctx.GetStub().PutState(id, valueJSON)
}

// DeleteAsset marks an asset as DELETED and restricts visibility (Soft Delete)
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	value, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	clientFullID, err := s.getClientFullIdentifier(ctx)
	if err != nil {
		return err
	}

	if value.Asset.OwnerID != clientFullID {
		return fmt.Errorf("only the owner can delete this asset")
	}

	txTimestamp, _ := ctx.GetStub().GetTxTimestamp()
	now := time.Unix(txTimestamp.Seconds, int64(txTimestamp.Nanos)).Format(time.RFC3339)

	// Soft Delete: Mark as DELETED and hide from public discovery
	value.Asset.Status = DeletedStatus
	value.Asset.View = PrivateView
	
	value.Audit = AuditMetadata{
		Action:    DeleteActionType,
		Actor:     clientFullID,
		Timestamp: now,
	}

	valueJSON, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = ctx.GetStub().SetEvent("DeleteAsset", valueJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	return ctx.GetStub().PutState(id, valueJSON)
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

		var value LedgerValue
		if len(response.Value) > 0 {
			var temp map[string]interface{}
			json.Unmarshal(response.Value, &temp)
			
			if _, ok := temp["asset"]; ok || temp["Asset"] != nil {
				json.Unmarshal(response.Value, &value)
			} else {
				// Legacy data handling
				var asset Asset
				json.Unmarshal(response.Value, &asset)
				value.Asset = asset
				value.Audit = AuditMetadata{Action: "LEGACY", Actor: "Unknown"}
			}
		}

		record := HistoryRecord{
			TxId:       response.TxId,
			Timestamp:  time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)),
			ActorID:    value.Audit.Actor,
			ActionType: value.Audit.Action,
			Value:      &value.Asset,
			IsDelete:   response.IsDelete,
		}
		
		if response.IsDelete {
			record.ActionType = DeleteActionType
		}

		records = append(records, record)
	}

	return records, nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*LedgerValue, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var values []*LedgerValue
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var value LedgerValue
		var temp map[string]interface{}
		json.Unmarshal(queryResponse.Value, &temp)
		
		if _, ok := temp["asset"]; ok || temp["Asset"] != nil {
			json.Unmarshal(queryResponse.Value, &value)
		} else {
			var asset Asset
			json.Unmarshal(queryResponse.Value, &asset)
			value.Asset = asset
			value.Audit = AuditMetadata{Action: "LEGACY", Actor: "Unknown"}
		}
		
		values = append(values, &value)
	}

	return values, nil
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

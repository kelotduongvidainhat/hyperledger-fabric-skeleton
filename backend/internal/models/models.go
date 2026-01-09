package models

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"uniqueIndex" json:"email"`
	Org       string    `json:"org"` // Org1MSP or Org2MSP
	Role      string    `gorm:"default:user" json:"role"`
	Status    string    `gorm:"default:PENDING" json:"status"` // PENDING, ACTIVE, BANNED
	DBStatus  string    `gorm:"default:Synced" json:"db_status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AssetAttachment struct {
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
	FileHash    string `json:"file_hash"`
	IpfsCID     string `json:"ipfs_cid"`
	StoragePath string `json:"storage_path"`
	StorageType string `json:"storage_type"`
}

type Asset struct {
	ID              string    `gorm:"primaryKey" json:"ID"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	OwnerID         string    `json:"ownerId"`
	ProposedOwnerID string    `json:"proposedOwnerId"`
	ImageURL        string    `json:"imageUrl"`
	ImageHash       string    `json:"imageHash"`
	Status          string          `json:"status"`
	View            string          `json:"view"`
	Attachment      AssetAttachment `gorm:"embedded;embeddedPrefix:attach_" json:"attachment"`
	// Metadata (Flattened for DB)
	LastUpdatedBy string    `json:"lastUpdatedBy"`
	LastUpdatedAt time.Time `json:"lastUpdatedAt"`
	Action        string    `json:"action"`
}

type AuditMetadata struct {
	Action    string `json:"action"`
	Actor     string `json:"actor"`
	Timestamp string `json:"timestamp"`
}

type LedgerValue struct {
	Asset Asset         `json:"asset"`
	Audit AuditMetadata `json:"audit"`
}

type HistoryRecord struct {
	TxId       string    `json:"txId"`
	Timestamp  time.Time `json:"timestamp"`
	ActorID    string    `json:"actorId"`
	ActionType string    `json:"actionType"`
	Value      *Asset    `json:"value"`
	IsDelete   bool      `json:"isDelete"`
}

type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    string    `gorm:"index" json:"user_id"` // Format: OrgMSP::Username
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Type      string    `json:"type"` // info, success, warning
	IsRead    bool      `gorm:"default:false" json:"is_read"`
	Link      string    `json:"link"`
	CreatedAt time.Time `json:"created_at"`
}

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

type Asset struct {
	ID              string    `gorm:"primaryKey" json:"ID"`
	Name            string    `json:"Name"`
	OwnerID         string    `json:"OwnerID"`
	Status          string    `json:"Status"`
	Description     string    `json:"Description"`
	ImageURL        string    `json:"ImageURL"`
	ImageHash       string    `json:"ImageHash"`
	ProposedOwnerID string    `json:"ProposedOwnerID"`
	View            string    `json:"View"`
	// Metadata (Flattened for DB)
	LastUpdatedBy string    `json:"LastUpdatedBy"`
	LastUpdatedAt time.Time `json:"LastUpdatedAt"`
	Action        string    `json:"Action"`
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

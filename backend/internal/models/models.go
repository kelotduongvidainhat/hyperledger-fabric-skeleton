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
	LastUpdatedBy   string    `json:"LastUpdatedBy"`
	LastUpdatedAt   time.Time `json:"LastUpdatedAt"`
	ProposedOwnerID string    `json:"ProposedOwnerID"`
	View            string    `json:"View"`
}

type HistoryRecord struct {
	TxId       string    `json:"TxId"`
	Timestamp  time.Time `json:"Timestamp"`
	ActorID    string    `json:"ActorID"`
	ActionType string    `json:"ActionType"`
	Value      *Asset    `json:"Value"`
	IsDelete   bool      `json:"IsDelete"`
}


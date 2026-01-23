package domain

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Email represents an email message stored in SQLite
type Email struct {
	// ID 是主键，手动设置 (Gmail ID)
	ID        string    `gorm:"primaryKey;column:id"` 
	ThreadID  string    `gorm:"index;column:thread_id"` 
	Subject   string    `gorm:"column:subject"`
	From      string    `gorm:"index;column:from_address"` // 'From' 是 SQL 关键字，最好改个名
	
	// 注意：数据库里存的是 JSON 字符串，但我们在业务代码里想用 []string
	// GORM 也可以用 serializer:json，但为了让你理解原理，这里演示手动转换
	ToJSON    string    `gorm:"column:to_list"` 
	
	Snippet   string    `gorm:"column:snippet"`
	BodyText  string    `gorm:"column:body_text"`
	
	Date      time.Time `gorm:"index;column:date"`
	
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"` // 软删除，相当于 Java 的 @SQLDelete(sql="UPDATE... SET deleted=true")
}

// TableName specifies the table name for GORM
func (Email) TableName() string {
	// TODO: Return the table name (e.g., "emails")
	return ""
}

// GetToList parses the To field (stored as JSON string) into a slice
func (e *Email) GetToList() ([]string, error) {
	var toList []string
	err := json.Unmarshal([]byte(e.ToJSON), &toList)
	return toList, err
}

// SetToList converts a slice to JSON string and stores in To field
func (e *Email) SetToList(recipients []string) error{
	bytes, err := json.Marshal(recipients)
	if err != nil{
		return err
	}
	e.ToJSON = string(bytes)
	return nil
}

// Chunk represents a text chunk from an email (for RAG)
type Chunk struct {
	// TODO: Add fields
	// ID, EmailID, Content, Position, TokenCount, Metadata
	// Use appropriate GORM tags
}

// TableName for Chunk
func (Chunk) TableName() string {
	// TODO: Return "chunks"
	return ""
}

// Embedding tracks which embeddings exist (actual vectors stored in Qdrant)
type Embedding struct {
	// TODO: Add fields
	// ID, EmailID, ChunkID, VectorID (ID in Qdrant), Metadata
}

// TableName for Embedding
func (Embedding) TableName() string {
	// TODO: Return "embeddings"
	return ""
}

// SyncMetadata tracks the last sync time with Gmail
type SyncMetadata struct {
	// TODO: Add fields
	// ID, LastSyncTime, EmailsCount
	// Hint: Use `gorm:"primaryKey;autoIncrement"` for ID
}

// TableName for SyncMetadata
func (SyncMetadata) TableName() string {
	// TODO: Return "sync_metadata"
	return ""
}

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
	return "emails"
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
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	EmailID   string    `gorm:"index;column:email_id"` // 外键逻辑关联
	
	Content   string    `gorm:"type:text;column:content"` // 真正被 embed 的文本
	
	Position  int       `gorm:"column:position"` // 在 email 中的顺序（第几个 chunk）
	TokenCnt  int       `gorm:"column:token_count"`
	
	Source    string    `gorm:"column:source"` 
	// body / subject / snippet
	
	CreatedAt time.Time
}

// TableName for Chunk
func (Chunk) TableName() string {
	return "chunks"
}

// Embedding tracks which embeddings exist (actual vectors stored in Qdrant)
type Embedding struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	
	ChunkID  uint      `gorm:"uniqueIndex;column:chunk_id"`
	EmailID  string    `gorm:"index;column:email_id"`

	VectorID string    `gorm:"uniqueIndex;column:vector_id"` 
	// Qdrant point ID
	
	Model    string    `gorm:"column:model"` 
	// text-embedding-3-small
	
	Dim      int       `gorm:"column:dimension"`
	
	CreatedAt time.Time
}


// TableName for Embedding
func (Embedding) TableName() string {
	return "embeddings"
}

// SyncMetadata tracks the last sync time with Gmail
type SyncMetadata struct {
	ID            uint      `gorm:"primaryKey;autoIncrement"`
	
	LastSyncTime  time.Time `gorm:"column:last_sync_time"`
	EmailsCount   int       `gorm:"column:emails_count"`
	
	CreatedAt     time.Time
	UpdatedAt     time.Time
}


// TableName for SyncMetadata
func (SyncMetadata) TableName() string {
	return "sync_metadata"
}

package vector

import "context"

// Repository defines operations for vector storage (embeddings)
type Repository interface {
	// Upsert inserts or updates vector embeddings
	// Points are the basic unit in Qdrant - each point has an ID, vector, and payload
	Upsert(ctx context.Context, points []*Point) error

	// Search finds similar vectors using cosine similarity
	Search(ctx context.Context, vector []float32, opts SearchOptions) ([]*SearchResult, error)

	// Delete removes vectors by their IDs
	Delete(ctx context.Context, pointIDs []string) error

	// DeleteByEmailID removes all vectors associated with an email
	// This uses payload filtering in Qdrant
	DeleteByEmailID(ctx context.Context, emailID string) error

	// CollectionInfo returns stats about the collection
	CollectionInfo(ctx context.Context) (*CollectionInfo, error)
}

// Point represents a vector point to store in Qdrant
type Point struct {
	ID       string                 // Unique point ID
	Vector   []float32              // The embedding vector (1536 dimensions)
	Payload  map[string]interface{} // Metadata (email_id, chunk_id, content, etc.)
}

// SearchOptions configures vector search
type SearchOptions struct {
	// 返回多少条（Top K）
	Limit int

	// 相似度阈值（0 ~ 1），低于这个直接丢掉
	ScoreThreshold float32

	// 可选：payload 过滤条件（Qdrant Filter）
	Filter map[string]interface{}
}


// SearchResult represents a search result from Qdrant
type SearchResult struct {
	// 命中的 point ID
	ID string

	// 相似度分数（cosine similarity）
	Score float32

	// 返回 payload（用于拿 chunk_id / email_id）
	Payload map[string]interface{}
}


type CollectionInfo struct {
	// collection 中向量总数
	VectorsCount int64

	// point 数量（一般等于 vectors）
	PointsCount int64

	// collection 状态：green / yellow / red
	Status string
}

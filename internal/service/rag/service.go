package rag

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/M1ngdaXie/go-local-rag-email/internal/domain"
	"github.com/M1ngdaXie/go-local-rag-email/internal/repository/vector"
	"github.com/M1ngdaXie/go-local-rag-email/internal/service/llm"
	"github.com/M1ngdaXie/go-local-rag-email/pkg/logger"
	"github.com/google/uuid"
)

// Service orchestrates chunking, embedding, and vector storage
type Service struct {
	vectorRepo vector.Repository
	llmService *llm.Service
	logger     logger.Logger
	chunkSize  int // Target tokens per chunk (~500)
	overlap    int // Overlap between chunks (~50)
}

// New creates a new RAG service
func New(vectorRepo vector.Repository, llmSvc *llm.Service, log logger.Logger) *Service {
	return &Service{
		vectorRepo: vectorRepo,
		llmService: llmSvc,
		logger:     log,
		chunkSize:  500, // ~500 tokens per chunk
		overlap:    50,  // ~50 token overlap
	}
}

// SearchResult represents a search result with email metadata
type SearchResult struct {
	EmailID string
	Score   float32
	Subject string
	From    string
}

// IndexEmail chunks an email, generates embeddings, and stores in Qdrant
func (s *Service) IndexEmail(ctx context.Context, email *domain.Email) (err error) {
    // 【防御 1】防止单个邮件的特殊数据导致整个同步进程崩溃
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("recovered from panic: %v", r)
        }
    }()

    content := prepareEmailContent(email)
    // 【清洗】修复非法 UTF-8，防止 Qdrant SDK 报错
    content = s.fixUTF8(content)
    cleanSubject := s.fixUTF8(email.Subject)

    if strings.TrimSpace(content) == "" {
        s.logger.Debug("Skipping email with empty content", "email_id", email.ID)
        return nil
    }

    chunks := s.chunkText(content)
    if len(chunks) == 0 {
        return nil
    }

    embeddings, err := s.llmService.GenerateEmbeddings(ctx, chunks)
    if err != nil {
        return fmt.Errorf("failed to generate embeddings: %w", err)
    }

    points := make([]*vector.Point, len(chunks))
    for i, chunk := range chunks {
        // 【幂等】使用确定性 ID，支持重复运行不重样
        id := uuid.NewMD5(uuid.Nil, []byte(email.ID+"_"+strconv.Itoa(i))).String()
        
        points[i] = &vector.Point{
            ID:     id,
            Vector: embeddings[i],
            Payload: map[string]interface{}{
                "email_id":       email.ID,
                "subject":        cleanSubject,
                "from":           email.From,
                "date":           email.Date.Format(time.RFC3339),
                "chunk_position": i,
                "content":        chunk, // 这里已经是 fixUTF8 过的
            },
        }
    }

    return s.vectorRepo.Upsert(ctx, points)
}

// 辅助函数：清洗无效字符
func (s *Service) fixUTF8(input string) string {
    if utf8.ValidString(input) {
        return input
    }
    // 过滤掉所有非法的 UTF-8 序列
    return strings.ToValidUTF8(input, "")
}

// IndexEmails indexes multiple emails (batch operation)
func (s *Service) IndexEmails(ctx context.Context, emails []*domain.Email) error {
	for i, email := range emails {
		s.logger.Info("Indexing email", "progress", fmt.Sprintf("%d/%d", i+1, len(emails)), "subject", email.Subject)

		if err := s.IndexEmail(ctx, email); err != nil {
			select {
			case <-ctx.Done():
    		return ctx.Err() // 如果用户按了 Ctrl+C，立刻停止后续所有邮件的处理
			default:
		}
			s.logger.Error("Failed to index email", "email_id", email.ID, "error", err)
			continue
		}
	}

	return nil
}

// Search performs semantic search and returns matching email IDs with scores
func (s *Service) Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if strings.TrimSpace(query) == "" {
		return nil, fmt.Errorf("empty query")
	}

	// Step 1: Generate query embedding
	queryVector, err := s.llmService.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Step 2: Search Qdrant
	searchResults, err := s.vectorRepo.Search(ctx, queryVector, vector.SearchOptions{
	    Limit: limit * 3,
	    ScoreThreshold: 0.1,
	})
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}
	s.logger.Debug("Qdrant search completed", "raw_results", len(searchResults))
	// Step 3: Deduplicate by email_id
	emailScores := make(map[string]SearchResult)
    for _, r := range searchResults {
        // 既然刚才已经检查过类型了，直接用变量名，既安全又简洁
        emailID,ok1 := r.Payload["email_id"].(string)
        subject,ok2 := r.Payload["subject"].(string)
        from, ok3 := r.Payload["from"].(string)
		if !ok1 || !ok2 || !ok3 {
            s.logger.Warn("Skipping result with missing fields", "point_id", r.ID)
            continue
        }
		if emailID == "<nil>" || subject == "<nil>" {
            s.logger.Warn("Skipping result: field is nil", "point_id", r.ID)
            continue
        }
		
        existing, exists := emailScores[emailID]
        if !exists || r.Score > existing.Score {
            emailScores[emailID] = SearchResult{
                EmailID: emailID,
                Score:   r.Score,
                Subject: subject, // 使用刚才断言出的变量
                From:    from,    // 使用刚才断言出的变量
            }
        }
    }

	// Step 4: Convert to slice and return top results
    finalResults := make([]SearchResult, 0, len(emailScores))
    for _, res := range emailScores {
        finalResults = append(finalResults, res)
    }

    // 按分数从高到低排序 (Descending)
    sort.Slice(finalResults, func(i, j int) bool {
        return finalResults[i].Score > finalResults[j].Score
    })

    // 截取到用户请求的 limit 数量
    if len(finalResults) > limit {
        finalResults = finalResults[:limit]
    }

    return finalResults, nil
}

// DeleteEmailIndex removes all vectors for an email
func (s *Service) DeleteEmailIndex(ctx context.Context, emailID string) error {
	return s.vectorRepo.DeleteByEmailID(ctx, emailID)
}

// chunkText splits text into overlapping chunks
func (s *Service) chunkText(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return []string{}
	}

	charChunkSize := s.chunkSize * 4 // ~4 chars per token
	charOverlap := s.overlap * 4

	// If text is short enough, return as single chunk
	if len(text) <= charChunkSize {
		return []string{text}
	}

	var chunks []string

	for i := 0; i < len(text); i += (charChunkSize - charOverlap) {
	    end := min(i+charChunkSize, len(text))
	    chunk := strings.TrimSpace(text[i:end])
	    if chunk != "" {
	        chunks = append(chunks, chunk)
	    }
	    if end == len(text) {
	        break
	    }
	}

	return chunks
}

// prepareEmailContent combines subject and body for indexing
func prepareEmailContent(email *domain.Email) string {
	var parts []string

	if email.Subject != "" {
		parts = append(parts, "Subject: "+email.Subject)
	}

	if email.BodyText != "" {
		parts = append(parts, email.BodyText)
	}

	if email.From != ""{
		parts = append(parts, "From " + email.From)
	}
	if email.Date.IsZero() == false {
    parts = append(parts, "Date: "+email.Date.Format("January 2, 2006"))
	}

	if email.ToJSON != "" {
    cleanTo := strings.NewReplacer("[", "", "]", "", "\"", "").Replace(email.ToJSON)
    parts = append(parts, "To: "+cleanTo)
}


	return strings.Join(parts, "\n\n")
}

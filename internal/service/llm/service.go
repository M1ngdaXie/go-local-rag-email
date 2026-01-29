package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/M1ngdaXie/go-local-rag-email/internal/config"
	openai "github.com/sashabaranov/go-openai"
)

// Service wraps OpenAI API for embedding generation
type Service struct {
	client *openai.Client
	model  openai.EmbeddingModel
}

// New creates a new LLM service
func New(cfg config.OpenAIConfig) (*Service, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	client := openai.NewClient(cfg.APIKey)

	return &Service{
		client: client,
		model:  openai.EmbeddingModel(cfg.EmbeddingModel), // e.g., "text-embedding-3-small"
	}, nil
}

// GenerateEmbedding generates a vector embedding for a single text input
func (s *Service) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	cleanText := strings.TrimSpace(text)
	if cleanText == "" {
		return nil, fmt.Errorf("empty text input: cannot generate embedding for empty string")
	}

	// Step 2: Create request
	req := openai.EmbeddingRequest{
		Model: openai.SmallEmbedding3, 
		Input: []string{cleanText},
	}

	// Step 3: Call API
	resp, err := s.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("openai embedding api error: %w", err)
	}
	// Step 4: Extract and return the embedding
	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned from openai")
	}

	return resp.Data[0].Embedding, nil
}

// GenerateEmbeddings generates vector embeddings for multiple texts (batched)
func (s *Service) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	// Step 1: 基础检查
	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	// Step 2: 预处理（防御性编程）
	validTexts := make([]string, len(texts))
	for i, t := range texts {
		trimmed := strings.TrimSpace(t)
		if trimmed == "" {
			validTexts[i] = "[empty_content]" 
		} else {
			validTexts[i] = trimmed
		}
	}

	// Step 3: 创建批量请求
	req := openai.EmbeddingRequest{
		Model: openai.SmallEmbedding3, // 使用 text-embedding-3-small
		Input: validTexts,             // 直接传入切片
	}

	// Step 4: 调用 API
	resp, err := s.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("openai batch embedding error: %w", err)
	}

	// Step 5: 按照原始顺序提取向量
	// OpenAI 保证 resp.Data[i] 对应 Input[i]
	embeddings := make([][]float32, len(resp.Data))
	for i, data := range resp.Data {
		embeddings[i] = data.Embedding
	}

	return embeddings, nil
}
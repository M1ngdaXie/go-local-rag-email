package vector

import (
	"context"
	"fmt"

	"github.com/M1ngdaXie/go-local-rag-email/internal/config"
	pkgLogger "github.com/M1ngdaXie/go-local-rag-email/pkg/logger"
	pb "github.com/qdrant/go-client/qdrant" // 别名 pb 方便引用
)

type qdrantRepo struct {
	client         *pb.Client
	collectionName string
	logger         pkgLogger.Logger
}

// NewQdrantRepository creates a new Qdrant-based vector repository
func NewQdrantRepository(client *pb.Client, cfg config.QdrantConfig, log pkgLogger.Logger) Repository {
	return &qdrantRepo{
		client:         client,
		collectionName: cfg.CollectionName,
		logger:         log,
	}
}

// Upsert inserts or updates vector points
func (r *qdrantRepo) Upsert(ctx context.Context, points []*Point) error {
	qdrantPoints := make([]*pb.PointStruct, len(points))
	
	for i, p := range points {
		// 1. ID 转换：使用辅助函数处理 UUID
		pointID := stringToPointID(p.ID)

		// 2. 构建 Qdrant Point
		qdrantPoints[i] = &pb.PointStruct{
			Id:      pointID,
			Vectors: pb.NewVectors(p.Vector...), // 注意这里解包 slice
			Payload: pb.NewValueMap(p.Payload),  // 自动把 Go map 转成 Qdrant Value
		}
	}

	// 3. 执行 Upsert
	wait := true // 等待写入落盘，确保一致性
	_, err := r.client.Upsert(ctx, &pb.UpsertPoints{
		CollectionName: r.collectionName,
		Wait:           &wait,
		Points:         qdrantPoints,
	})

	if err != nil {
		// 记录详细错误方便调试
		r.logger.Error("Failed to upsert to Qdrant", "error", err)
		return fmt.Errorf("qdrant upsert failed: %w", err)
	}

	r.logger.Debug("Upserted vectors", "count", len(points))
	return nil
}

// Search finds similar vectors
func (r *qdrantRepo) Search(ctx context.Context, vec []float32, opts SearchOptions) ([]*SearchResult, error) {
	// 1. 构建搜索请求 (使用 Query API)
	limit := uint64(opts.Limit)

	queryPoints := &pb.QueryPoints{
		CollectionName: r.collectionName,
		Query:          pb.NewQuery(vec...),
		Limit:          &limit,
		ScoreThreshold: &opts.ScoreThreshold,
		WithPayload:    pb.NewWithPayload(true),
	}

	// TODO: 如果 opts.Filter 不为空，这里需要构建 Filter (比较复杂，MVP 先跳过)

	// 2. 执行搜索
	resp, err := r.client.Query(ctx, queryPoints)
	if err != nil {
		return nil, fmt.Errorf("qdrant query failed: %w", err)
	}

	// 3. 结果转换 (Qdrant -> Domain)
	results := make([]*SearchResult, len(resp))
	for i, item := range resp {
		results[i] = &SearchResult{
			ID:      item.Id.GetUuid(),
			Score:   item.Score,
			Payload: make(map[string]interface{}),
		}

		// 提取 payload 中的字段
		if item.Payload != nil {
			if val, ok := item.Payload["email_id"]; ok {
				if strVal := val.GetStringValue(); strVal != "" {
					results[i].Payload["email_id"] = strVal
				}
			}
			if val, ok := item.Payload["content"]; ok {
				if strVal := val.GetStringValue(); strVal != "" {
					results[i].Payload["content"] = strVal
				}
			}
			if val, ok := item.Payload["chunk_id"]; ok {
				if strVal := val.GetStringValue(); strVal != "" {
					results[i].Payload["chunk_id"] = strVal
				}
			}
		}
	}

	r.logger.Debug("Vector search completed", "hits", len(results))
	return results, nil
}

// Delete removes vectors by IDs
func (r *qdrantRepo) Delete(ctx context.Context, pointIDs []string) error {
	// 1. 转换 ID
	ids := make([]*pb.PointId, len(pointIDs))
	for i, id := range pointIDs {
		ids[i] = stringToPointID(id)
	}

	// 2. 执行删除 (使用 PointsSelectorOneOf)
	wait := true
	_, err := r.client.Delete(ctx, &pb.DeletePoints{
		CollectionName: r.collectionName,
		Wait:           &wait,
		Points: &pb.PointsSelector{
			PointsSelectorOneOf: &pb.PointsSelector_Points{
				Points: &pb.PointsIdsList{Ids: ids},
			},
		},
	})

	if err != nil {
		return fmt.Errorf("qdrant delete failed: %w", err)
	}

	return nil
}

// DeleteByEmailID removes all vectors for an email using payload filter
func (r *qdrantRepo) DeleteByEmailID(ctx context.Context, emailID string) error {
	wait := true

	// 1. 构建 Filter: payload["email_id"] == emailID
	filter := &pb.Filter{
		Must: []*pb.Condition{
			pb.NewMatchKeyword("email_id", emailID),
		},
	}

	// 2. 执行按条件删除 (使用 PointsSelectorOneOf)
	_, err := r.client.Delete(ctx, &pb.DeletePoints{
		CollectionName: r.collectionName,
		Wait:           &wait,
		Points: &pb.PointsSelector{
			PointsSelectorOneOf: &pb.PointsSelector_Filter{
				Filter: filter,
			},
		},
	})

	if err != nil {
		return fmt.Errorf("qdrant delete_by_filter failed: %w", err)
	}

	r.logger.Info("Deleted vector chunks for email", "email_id", emailID)
	return nil
}

// CollectionInfo returns collection statistics
func (r *qdrantRepo) CollectionInfo(ctx context.Context) (*CollectionInfo, error) {
	// 1. 获取信息 (直接传 collection name)
	info, err := r.client.GetCollectionInfo(ctx, r.collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection info: %w", err)
	}

	// 2. 转换状态
	status := "unknown"
	if info != nil {
		status = info.Status.String()
	}

	// 3. 处理指针类型的计数字段
	var pointsCount int64
	if info.PointsCount != nil {
		pointsCount = int64(*info.PointsCount)
	}

	// VectorsCount 可能在 IndexedVectorsCount 中
	var vectorsCount int64
	if info.IndexedVectorsCount != nil {
		vectorsCount = int64(*info.IndexedVectorsCount)
	} else {
		vectorsCount = pointsCount // 如果没有索引向量计数，使用点数
	}

	return &CollectionInfo{
		VectorsCount: vectorsCount,
		PointsCount:  pointsCount,
		Status:       status,
	}, nil
}

// Helper: Convert string to Qdrant UUID PointID
func stringToPointID(s string) *pb.PointId {
	return &pb.PointId{
		PointIdOptions: &pb.PointId_Uuid{Uuid: s},
	}
}
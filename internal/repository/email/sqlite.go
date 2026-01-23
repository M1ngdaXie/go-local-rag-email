package email

import (
	"context"
	"fmt"

	"github.com/M1ngdaXie/go-local-rag-email/internal/domain"
	"github.com/M1ngdaXie/go-local-rag-email/pkg/logger"
	"gorm.io/gorm"
)

type sqliteRepo struct {
	db     *gorm.DB
	logger logger.Logger
}

// NewSQLiteRepository creates a new SQLite-based email repository
func NewSQLiteRepository(db *gorm.DB, log logger.Logger) Repository {
	return &sqliteRepo{
		db:     db,
		logger: log,
	}
}

// Create stores a new email
func (r *sqliteRepo) Create(ctx context.Context, email *domain.Email) error {
	// TODO: Use GORM to create the email
	result := r.db.WithContext(ctx).Create(email)
	if result.Error != nil{
		return result.Error
	}
	r.logger.Debug("Created email", "id", email.ID)
	return nil
}

// Get retrieves an email by ID
func (r *sqliteRepo) Get(ctx context.Context, id string) (*domain.Email, error) {
	// TODO: Query the database for an email with the given ID
	// Hint: var email domain.Email
	//       err := r.db.WithContext(ctx).Where("id = ?", id).First(&email).Error
	// Check for gorm.ErrRecordNotFound to return a friendly error

	return nil, fmt.Errorf("TODO: Implement Get")
}

// List retrieves emails with filters and pagination
func (r *sqliteRepo) List(ctx context.Context, filter Filter, page Pagination) ([]*domain.Email, error) {
	// TODO: Build a query with filters
	// Hint: var emails []*domain.Email
	//       query := r.db.WithContext(ctx)
	//
	// Apply filters (if not empty):
	// if filter.From != "" {
	//     query = query.Where("from LIKE ?", "%"+filter.From+"%")
	// }
	// if filter.DateFrom != nil {
	//     query = query.Where("date >= ?", filter.DateFrom)
	// }
	//
	// Apply pagination:
	// if page.Limit > 0 {
	//     query = query.Limit(page.Limit)
	// }
	// if page.Offset > 0 {
	//     query = query.Offset(page.Offset)
	// }
	//
	// Order by date descending:
	// query = query.Order("date DESC")
	//
	// Execute query:
	// err := query.Find(&emails).Error

	return nil, fmt.Errorf("TODO: Implement List")
}

// Count returns the total number of emails matching the filter
func (r *sqliteRepo) Count(ctx context.Context, filter Filter) (int64, error) {
	// TODO: Count emails matching the filter
	// Hint: var count int64
	//       query := r.db.WithContext(ctx).Model(&domain.Email{})
	//       Apply the same filters as List()
	//       err := query.Count(&count).Error

	return 0, fmt.Errorf("TODO: Implement Count")
}

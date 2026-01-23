package email

import (
	"context"
	"time"

	"github.com/M1ngdaXie/go-local-rag-email/internal/domain"
)

// Repository defines operations for email storage
type Repository interface {
	// Create stores a new email in the database
	Create(ctx context.Context, email *domain.Email) error

	// Get retrieves an email by ID
	Get(ctx context.Context, id string) (*domain.Email, error)

	// List retrieves emails with filters and pagination
	List(ctx context.Context, filter Filter, page Pagination) ([]*domain.Email, error)

	// Count returns the total number of emails matching the filter
	Count(ctx context.Context, filter Filter) (int64, error)
}

// Filter holds criteria for filtering emails
type Filter struct {
	// TODO: Add filter fields
	// Hint: From string, Labels []string, DateFrom *time.Time, DateTo *time.Time
}

// Pagination holds offset and limit for paging
type Pagination struct {
	// TODO: Add pagination fields
	// Hint: Offset int, Limit int
}

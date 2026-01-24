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
	var email domain.Email
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&email).Error
	if err != nil {                                                                                                                                                                                                    
      if err == gorm.ErrRecordNotFound {                                                                                                                                                                             
          return nil, fmt.Errorf("email not found: %s", id)                                                                                                                                                          
      }                                                                                                                                                                                                              
      return nil, fmt.Errorf("failed to get email: %w", err)                                                                                                                                                         
  } 
	r.logger.Debug("finding the email", "id", email.ID)
	return &email, nil
}

// List retrieves emails with filters and pagination
func (r *sqliteRepo) List(ctx context.Context, filter Filter, page Pagination) ([]*domain.Email, error) {
	var emails []*domain.Email
	query := r.buildFilter(ctx, filter)
	if page.Limit > 0 {                                                                                                                                                                                            
          query = query.Limit(page.Limit)                                                                                                                                                                            
      }                                                                                                                                                                                                              
      if page.Offset > 0 {                                                                                                                                                                                           
          query = query.Offset(page.Offset)                                                                                                                                                                          
      }   
	// Order by date descending:
	query = query.Order("date DESC")
	
	// Execute query:
	err := query.Find(&emails).Error
	if err != nil {
        return nil, err
    }
	return emails, nil
}
func (r *sqliteRepo) buildFilter(ctx context.Context, filter Filter) *gorm.DB {
    query := r.db.WithContext(ctx).Model(&domain.Email{})

    if filter.From != "" {
        // 记得要跟 domain tag 里的 column 名字一致
        query = query.Where("from_address LIKE ?", "%"+filter.From+"%")
    }
    if filter.DateFrom != nil {
        query = query.Where("date >= ?", *filter.DateFrom)
    }
    
    return query
}
// Count returns the total number of emails matching the filter
func (r *sqliteRepo) Count(ctx context.Context, filter Filter) (int64, error) {
	// TODO: Count emails matching the filter
	var count int64
	      query := r.buildFilter(ctx, filter)
	    //   Apply the same filters as List()
	      err := query.Count(&count).Error
		  if err != nil{
			return count, err
		  }
	return count, nil
}

package repositories

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/brayanzuritadev/StockManager/internal/models/domain"
)

// IProductImageRepository defines the persistence contract for product images.
type IProductImageRepository interface {
	GetByProductID(productID int) ([]domain.ProductImage, error)
	GetByProductIDs(productIDs []int) (map[int][]domain.ProductImage, error)
	GetOne(productID, imageID int) (*domain.ProductImage, error)
	Add(productID int, colorID *int, url, publicID string, sortOrder int) (*domain.ProductImage, error)
	UpdateColorID(productID, imageID int, colorID *int) (*domain.ProductImage, error)
	Delete(productID, imageID int) error
	DeleteAllByProductID(productID int) error
}

// ProductImageRepository is the PostgreSQL implementation of IProductImageRepository.
type ProductImageRepository struct {
	db *sql.DB
}

func NewProductImageRepository(db *sql.DB) IProductImageRepository {
	return &ProductImageRepository{db: db}
}

func (r *ProductImageRepository) GetByProductID(productID int) ([]domain.ProductImage, error) {
	rows, err := r.db.Query(`
		SELECT id, product_id, color_id, url, public_id, sort_order, created_at
		FROM product_images
		WHERE product_id = $1
		ORDER BY sort_order, id`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var images []domain.ProductImage
	for rows.Next() {
		var img domain.ProductImage
		if err := rows.Scan(&img.ID, &img.ProductID, &img.ColorID, &img.URL, &img.PublicID, &img.SortOrder, &img.CreatedAt); err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, nil
}

func (r *ProductImageRepository) GetOne(productID, imageID int) (*domain.ProductImage, error) {
	img := &domain.ProductImage{}
	err := r.db.QueryRow(`
		SELECT id, product_id, color_id, url, public_id, sort_order, created_at
		FROM product_images
		WHERE id = $1 AND product_id = $2`, imageID, productID).
		Scan(&img.ID, &img.ProductID, &img.ColorID, &img.URL, &img.PublicID, &img.SortOrder, &img.CreatedAt)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func (r *ProductImageRepository) UpdateColorID(productID, imageID int, colorID *int) (*domain.ProductImage, error) {
	img := &domain.ProductImage{}
	err := r.db.QueryRow(`
		UPDATE product_images
		SET color_id = $1
		WHERE id = $2 AND product_id = $3
		RETURNING id, product_id, color_id, url, public_id, sort_order, created_at`,
		colorID, imageID, productID).
		Scan(&img.ID, &img.ProductID, &img.ColorID, &img.URL, &img.PublicID, &img.SortOrder, &img.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("image not found")
	}
	return img, nil
}

// GetByProductIDs fetches all images for a list of product IDs in one query.
// Returns a map[productID][]ProductImage.
func (r *ProductImageRepository) GetByProductIDs(productIDs []int) (map[int][]domain.ProductImage, error) {
	result := make(map[int][]domain.ProductImage)
	if len(productIDs) == 0 {
		return result, nil
	}

	// Build $1,$2,... placeholders
	placeholders := make([]string, len(productIDs))
	args := make([]interface{}, len(productIDs))
	for i, id := range productIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := `SELECT id, product_id, color_id, url, public_id, sort_order, created_at
		FROM product_images
		WHERE product_id IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY product_id, sort_order, id`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var img domain.ProductImage
		if err := rows.Scan(&img.ID, &img.ProductID, &img.ColorID, &img.URL, &img.PublicID, &img.SortOrder, &img.CreatedAt); err != nil {
			return nil, err
		}
		result[img.ProductID] = append(result[img.ProductID], img)
	}
	return result, nil
}

func (r *ProductImageRepository) Add(productID int, colorID *int, url, publicID string, sortOrder int) (*domain.ProductImage, error) {
	img := &domain.ProductImage{}
	err := r.db.QueryRow(`
		INSERT INTO product_images (product_id, color_id, url, public_id, sort_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, product_id, color_id, url, public_id, sort_order, created_at`,
		productID, colorID, url, publicID, sortOrder).
		Scan(&img.ID, &img.ProductID, &img.ColorID, &img.URL, &img.PublicID, &img.SortOrder, &img.CreatedAt)
	return img, err
}

func (r *ProductImageRepository) Delete(productID, imageID int) error {
	res, err := r.db.Exec(`
		DELETE FROM product_images WHERE id = $1 AND product_id = $2`, imageID, productID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("image not found")
	}
	return nil
}

func (r *ProductImageRepository) DeleteAllByProductID(productID int) error {
	_, err := r.db.Exec(`DELETE FROM product_images WHERE product_id = $1`, productID)
	return err
}

package services

import (
	"fmt"

	"github.com/brayanzuritadev/StockManager/internal/models/domain"
	"github.com/brayanzuritadev/StockManager/internal/repositories"
)

// IInventoryMovementService defines the business logic for inventory movements.
type IInventoryMovementService interface {
	GetAll() ([]domain.InventoryMovement, error)
	GetByID(id int) (*domain.InventoryMovement, error)
	GetByVariantID(variantID int) ([]domain.InventoryMovement, error)
	Create(variantID int, mvType string, quantity int, unitCost *float64, refType *string, refID *int) (*domain.InventoryMovement, error)
}

type InventoryMovementService struct {
	repo        repositories.IInventoryMovementRepository
	variantRepo repositories.IProductVariantRepository
}

func NewInventoryMovementService(
	repo repositories.IInventoryMovementRepository,
	variantRepo repositories.IProductVariantRepository,
) IInventoryMovementService {
	return &InventoryMovementService{repo: repo, variantRepo: variantRepo}
}

func (s *InventoryMovementService) GetAll() ([]domain.InventoryMovement, error) {
	return s.repo.GetAll()
}

func (s *InventoryMovementService) GetByID(id int) (*domain.InventoryMovement, error) {
	return s.repo.GetByID(id)
}

func (s *InventoryMovementService) GetByVariantID(variantID int) ([]domain.InventoryMovement, error) {
	return s.repo.GetByVariantID(variantID)
}

func (s *InventoryMovementService) Create(variantID int, mvType string, quantity int, unitCost *float64, refType *string, refID *int) (*domain.InventoryMovement, error) {
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be greater than 0")
	}
	if mvType != "IN" && mvType != "OUT" {
		return nil, fmt.Errorf("type must be IN or OUT")
	}

	v, err := s.variantRepo.GetByID(variantID)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, fmt.Errorf("variant %d not found", variantID)
	}

	if mvType == "OUT" && v.Stock < quantity {
		return nil, fmt.Errorf("insufficient stock (available: %d, requested: %d)", v.Stock, quantity)
	}

	movement, err := s.repo.Create(variantID, mvType, quantity, unitCost, refType, refID)
	if err != nil {
		return nil, err
	}

	// Adjust variant stock
	newStock := v.Stock + quantity
	if mvType == "OUT" {
		newStock = v.Stock - quantity
	}
	_ = s.variantRepo.UpdateStock(variantID, newStock)

	return movement, nil
}

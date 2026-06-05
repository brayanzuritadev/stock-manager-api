package services

import (
	"fmt"

	"github.com/brayanzuritadev/StockManager/internal/models/domain"
	"github.com/brayanzuritadev/StockManager/internal/repositories"
)

type ISaleService interface {
	GetAll() ([]domain.Sale, error)
	GetByID(id int) (*domain.Sale, error)
	Create(userID *int, channel *string, status *string, totalDiscount float64, notes *string, items []SaleItemInput) (*domain.Sale, error)
	UpdateStatus(id int, status string) error
	Delete(id int) error
}

type SaleItemInput struct {
	ProductVariantID int
	Quantity         int
	UnitPrice        float64
	Discount         float64
}

type SaleService struct {
	repo        repositories.ISaleRepository
	variantRepo repositories.IProductVariantRepository
	invRepo     repositories.IInventoryMovementRepository
}

func NewSaleService(
	repo repositories.ISaleRepository,
	variantRepo repositories.IProductVariantRepository,
	invRepo repositories.IInventoryMovementRepository,
) ISaleService {
	return &SaleService{
		repo:        repo,
		variantRepo: variantRepo,
		invRepo:     invRepo,
	}
}

func (s *SaleService) GetAll() ([]domain.Sale, error) {
	return s.repo.GetAll()
}

func (s *SaleService) GetByID(id int) (*domain.Sale, error) {
	sale, err := s.repo.GetByID(id)
	if err != nil || sale == nil {
		return sale, err
	}
	items, err := s.repo.GetItems(id)
	if err != nil {
		return nil, err
	}
	sale.Items = items
	return sale, nil
}

func (s *SaleService) Create(userID *int, channel *string, status *string, totalDiscount float64, notes *string, items []SaleItemInput) (*domain.Sale, error) {
	if len(items) == 0 {
		return nil, fmt.Errorf("sale must have at least one item")
	}

	// Verify stock availability
	for _, it := range items {
		v, err := s.variantRepo.GetByID(it.ProductVariantID)
		if err != nil {
			return nil, err
		}
		if v == nil {
			return nil, fmt.Errorf("variant %d not found", it.ProductVariantID)
		}
		if v.Stock < it.Quantity {
			return nil, fmt.Errorf("stock insuficiente para \"" + v.ProductName + "\" talla " + v.SizeName + " / " + v.ColorName +
				fmt.Sprintf(" (disponible: %d, solicitado: %d)", v.Stock, it.Quantity))
		}
	}

	// Compute total
	var total float64
	for _, it := range items {
		total += (it.UnitPrice - it.Discount) * float64(it.Quantity)
	}
	total -= totalDiscount

	// Create the sale record
	sale, err := s.repo.Create(userID, total, totalDiscount, channel, status, notes)
	if err != nil {
		return nil, err
	}

	saleIDRef := &sale.ID
	refType := "sale"

	for _, it := range items {
		// Record sale item
		if _, err := s.repo.AddItem(sale.ID, it.ProductVariantID, it.Quantity, it.UnitPrice, it.Discount); err != nil {
			return nil, err
		}

		// Record inventory OUT movement
		if _, err := s.invRepo.Create(it.ProductVariantID, "OUT", it.Quantity, &it.UnitPrice, &refType, saleIDRef); err != nil {
			return nil, err
		}

		// Decrement variant stock
		v, _ := s.variantRepo.GetByID(it.ProductVariantID)
		if v != nil {
			_ = s.variantRepo.UpdateStock(it.ProductVariantID, v.Stock-it.Quantity)
		}
	}

	return s.GetByID(sale.ID)
}

func (s *SaleService) UpdateStatus(id int, status string) error {
	allowed := map[string]bool{"pending": true, "paid": true, "cancelled": true, "refunded": true}
	if !allowed[status] {
		return fmt.Errorf("invalid status '%s'", status)
	}
	return s.repo.UpdateStatus(id, status)
}

func (s *SaleService) Delete(id int) error {
	return s.repo.Delete(id)
}

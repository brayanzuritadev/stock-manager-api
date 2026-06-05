package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/brayanzuritadev/StockManager/internal/models/domain"
	"github.com/brayanzuritadev/StockManager/internal/repositories"
)

// ICartService defines the business logic contract for carts.
type ICartService interface {
	GetAll() ([]domain.Cart, error)
	GetByID(id int) (*domain.Cart, error)
	GetBySharedLink(link string) (*domain.Cart, error)
	Create(userID *int) (*domain.Cart, error)
	UpdateStatus(id int, status string) (*domain.Cart, error)
	GenerateSharedLink(cartID int) (*domain.Cart, error)
	UpdateNotesByLink(link string, notes string) (*domain.Cart, error)
	Delete(id int) error
	GetItems(cartID int) ([]domain.CartItem, error)
	AddItem(cartID, productVariantID, quantity int, unitPrice, discount float64) (*domain.CartItem, error)
	UpdateItem(cartID, itemID, quantity int, discount float64) (*domain.CartItem, error)
	DeleteItem(cartID, itemID int) error
}

// CartService is the concrete implementation of ICartService.
type CartService struct {
	repo repositories.ICartRepository
}

func NewCartService(repo repositories.ICartRepository) ICartService {
	return &CartService{repo: repo}
}

func (s *CartService) GetAll() ([]domain.Cart, error) {
	return s.repo.GetAll()
}

func (s *CartService) GetByID(id int) (*domain.Cart, error) {
	cart, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if cart == nil {
		return nil, nil
	}
	items, err := s.repo.GetItems(id)
	if err != nil {
		return nil, err
	}
	cart.Items = items
	return cart, nil
}

func (s *CartService) GetBySharedLink(link string) (*domain.Cart, error) {
	cart, err := s.repo.GetBySharedLink(link)
	if err != nil {
		return nil, err
	}
	if cart == nil {
		return nil, nil
	}
	items, err := s.repo.GetItems(cart.ID)
	if err != nil {
		return nil, err
	}
	cart.Items = items
	return cart, nil
}

func (s *CartService) Create(userID *int) (*domain.Cart, error) {
	return s.repo.Create(userID)
}

func (s *CartService) UpdateStatus(id int, status string) (*domain.Cart, error) {
	allowed := map[string]bool{"pending": true, "completed": true, "cancelled": true}
	if !allowed[status] {
		return nil, fmt.Errorf("invalid status '%s'", status)
	}
	cart, err := s.repo.UpdateStatus(id, status)
	if err != nil {
		return nil, err
	}
	return cart, nil
}

func (s *CartService) GenerateSharedLink(cartID int) (*domain.Cart, error) {
	cart, err := s.repo.GetByID(cartID)
	if err != nil {
		return nil, err
	}
	if cart == nil {
		return nil, fmt.Errorf("cart not found")
	}
	if cart.SharedLink != nil && *cart.SharedLink != "" {
		return cart, nil
	}
	link, err := generateToken(16)
	if err != nil {
		return nil, fmt.Errorf("failed to generate shared link: %w", err)
	}
	return s.repo.SetSharedLink(cartID, link)
}

func (s *CartService) UpdateNotesByLink(link string, notes string) (*domain.Cart, error) {
	return s.repo.UpdateNotesByLink(link, notes)
}

func (s *CartService) Delete(id int) error {
	return s.repo.SoftDelete(id)
}

// ──────────────────────────
// Cart Items
// ──────────────────────────

func (s *CartService) GetItems(cartID int) ([]domain.CartItem, error) {
	cart, err := s.repo.GetByID(cartID)
	if err != nil {
		return nil, err
	}
	if cart == nil {
		return nil, fmt.Errorf("cart not found")
	}
	return s.repo.GetItems(cartID)
}

func (s *CartService) AddItem(cartID, productVariantID, quantity int, unitPrice, discount float64) (*domain.CartItem, error) {
	cart, err := s.repo.GetByID(cartID)
	if err != nil {
		return nil, err
	}
	if cart == nil {
		return nil, fmt.Errorf("cart not found")
	}
	if cart.Status == "completed" || cart.Status == "cancelled" {
		return nil, fmt.Errorf("cannot add items to a %s cart", cart.Status)
	}
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be greater than 0")
	}
	return s.repo.AddItem(cartID, productVariantID, quantity, unitPrice, discount)
}

func (s *CartService) UpdateItem(cartID, itemID, quantity int, discount float64) (*domain.CartItem, error) {
	cart, err := s.repo.GetByID(cartID)
	if err != nil {
		return nil, err
	}
	if cart == nil {
		return nil, fmt.Errorf("cart not found")
	}
	if cart.Status == "completed" || cart.Status == "cancelled" {
		return nil, fmt.Errorf("cannot modify items of a %s cart", cart.Status)
	}
	if quantity <= 0 {
		return nil, fmt.Errorf("quantity must be greater than 0")
	}
	item, err := s.repo.UpdateItem(cartID, itemID, quantity, discount)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, fmt.Errorf("item not found")
	}
	return item, nil
}

func (s *CartService) DeleteItem(cartID, itemID int) error {
	cart, err := s.repo.GetByID(cartID)
	if err != nil {
		return err
	}
	if cart == nil {
		return fmt.Errorf("cart not found")
	}
	return s.repo.DeleteItem(cartID, itemID)
}

// ──────────────────────────
// Helpers
// ──────────────────────────

func generateToken(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

package services

import (
	"errors"

	"github.com/brayanzuritadev/StockManager/internal/models/domain"
	"github.com/brayanzuritadev/StockManager/internal/models/dto"
	"github.com/brayanzuritadev/StockManager/internal/repositories"
)

// ICategoryService defines the business logic contract for categories.
type ICategoryService interface {
	GetAll() ([]dto.CategoryResponse, error)
	GetByID(id int) (*dto.CategoryResponse, error)
	Create(req dto.CreateCategoryRequest) (*dto.CategoryResponse, error)
	Update(id int, req dto.UpdateCategoryRequest) (*dto.CategoryResponse, error)
	Delete(id int) error
}

// CategoryService is the concrete implementation of ICategoryService.
type CategoryService struct {
	repo repositories.ICategoryRepository
}

func NewCategoryService(repo repositories.ICategoryRepository) ICategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) GetAll() ([]dto.CategoryResponse, error) {
	categories, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	result := make([]dto.CategoryResponse, 0, len(categories))
	for _, c := range categories {
		result = append(result, toCategoryResponse(c))
	}
	return result, nil
}

func (s *CategoryService) GetByID(id int) (*dto.CategoryResponse, error) {
	c, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, nil
	}
	res := toCategoryResponse(*c)
	return &res, nil
}

func (s *CategoryService) Create(req dto.CreateCategoryRequest) (*dto.CategoryResponse, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}
	c, err := s.repo.Create(domain.Category{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, err
	}
	res := toCategoryResponse(*c)
	return &res, nil
}

func (s *CategoryService) Update(id int, req dto.UpdateCategoryRequest) (*dto.CategoryResponse, error) {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Description != nil {
		existing.Description = req.Description
	}
	updated, err := s.repo.Update(id, *existing)
	if err != nil {
		return nil, err
	}
	res := toCategoryResponse(*updated)
	return &res, nil
}

func (s *CategoryService) Delete(id int) error {
	return s.repo.SoftDelete(id)
}

func toCategoryResponse(c domain.Category) dto.CategoryResponse {
	return dto.CategoryResponse{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
	}
}

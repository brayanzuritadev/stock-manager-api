package services

import (
	"fmt"

	"github.com/brayanzuritadev/StockManager/internal/models/domain"
	"github.com/brayanzuritadev/StockManager/internal/models/dto"
	"github.com/brayanzuritadev/StockManager/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

// IUserService defines the business logic contract for users.
type IUserService interface {
	GetAll() ([]dto.UserResponse, error)
	GetByID(id int) (*dto.UserResponse, error)
	Create(req dto.CreateUserRequest) (*dto.UserResponse, error)
	Update(id int, req dto.UpdateUserRequest) (*dto.UserResponse, error)
	Delete(id int) error
}

// UserService is the concrete implementation of IUserService.
type UserService struct {
	repo repositories.IUserRepository
}

func NewUserService(repo repositories.IUserRepository) IUserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetAll() ([]dto.UserResponse, error) {
	users, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	result := make([]dto.UserResponse, 0, len(users))
	for _, u := range users {
		result = append(result, toUserResponse(u))
	}
	return result, nil
}

func (s *UserService) GetByID(id int) (*dto.UserResponse, error) {
	u, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, nil
	}
	res := toUserResponse(*u)
	return &res, nil
}

func (s *UserService) Create(req dto.CreateUserRequest) (*dto.UserResponse, error) {
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("name, email and password are required")
	}

	existing, err := s.repo.GetByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("email already in use")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	u, err := s.repo.Create(req.Name, req.Email, string(hash), req.Phone)
	if err != nil {
		return nil, err
	}
	res := toUserResponse(*u)
	return &res, nil
}

func (s *UserService) Update(id int, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	existing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	name := existing.Name
	if req.Name != "" {
		name = req.Name
	}
	phone := req.Phone
	if phone == nil {
		phone = existing.Phone
	}

	u, err := s.repo.Update(id, name, phone)
	if err != nil {
		return nil, err
	}
	res := toUserResponse(*u)
	return &res, nil
}

func (s *UserService) Delete(id int) error {
	return s.repo.SoftDelete(id)
}

// ──────────────────────────
// mappers
// ──────────────────────────

func toUserResponse(u domain.User) dto.UserResponse {
	return dto.UserResponse{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
		Phone: u.Phone,
	}
}

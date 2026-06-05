package services

import (
	"context"
	"fmt"
	"log"

	"github.com/brayanzuritadev/StockManager/internal/models/domain"
	"github.com/brayanzuritadev/StockManager/internal/models/dto"
	"github.com/brayanzuritadev/StockManager/internal/repositories"
)

// ICloudinaryDeleter can remove a file from Cloudinary by its public ID.
type ICloudinaryDeleter interface {
	DeleteFile(ctx context.Context, publicID string) error
}

// IProductService defines the business logic contract for products.
type IProductService interface {
	GetAll(onlyActive bool) ([]dto.ProductResponse, error)
	GetActiveWithStock() ([]dto.ProductResponse, error)
	GetByID(id int) (*dto.ProductResponse, error)
	GetByCategoryID(categoryID int) ([]dto.ProductResponse, error)
	Create(req dto.CreateProductRequest) (*dto.ProductResponse, error)
	Update(id int, req dto.UpdateProductRequest) (*dto.ProductResponse, error)
	Delete(id int) error
	GetImages(productID int) ([]dto.ProductImageResponse, error)
	AddImage(productID int, req dto.AddProductImageRequest) (*dto.ProductImageResponse, error)
	UpdateImageColor(productID, imageID int, colorID *int) (*dto.ProductImageResponse, error)
	DeleteImage(productID, imageID int) error
	GetVariants(productID int) ([]dto.ProductVariantResponse, error)
	AddVariant(productID int, req dto.CreateProductVariantRequest) (*dto.ProductVariantResponse, error)
	UpdateVariant(productID, variantID int, req dto.UpdateProductVariantRequest) (*dto.ProductVariantResponse, error)
	DeleteVariant(productID, variantID int) error
}

// ProductService is the concrete implementation of IProductService.
type ProductService struct {
	repo          repositories.IProductRepository
	imageRepo     repositories.IProductImageRepository
	variantRepo   repositories.IProductVariantRepository
	inventoryRepo repositories.IInventoryMovementRepository
	cld           ICloudinaryDeleter // may be nil
}

func NewProductService(
	repo repositories.IProductRepository,
	imageRepo repositories.IProductImageRepository,
	variantRepo repositories.IProductVariantRepository,
	inventoryRepo repositories.IInventoryMovementRepository,
	cld ICloudinaryDeleter,
) IProductService {
	return &ProductService{repo: repo, imageRepo: imageRepo, variantRepo: variantRepo, inventoryRepo: inventoryRepo, cld: cld}
}

func (s *ProductService) GetAll(onlyActive bool) ([]dto.ProductResponse, error) {
	products, err := s.repo.GetAll(onlyActive)
	if err != nil {
		return nil, err
	}
	imagesMap, err := s.loadImagesForProducts(products)
	if err != nil {
		return nil, err
	}
	ids := make([]int, len(products))
	for i, p := range products {
		ids[i] = p.ID
	}
	variantsMap, err := s.variantRepo.GetByProductIDs(ids)
	if err != nil {
		return nil, err
	}
	result := make([]dto.ProductResponse, 0, len(products))
	for _, p := range products {
		p.Images = imagesMap[p.ID]
		p.Variants = variantsMap[p.ID]
		result = append(result, toProductResponse(p))
	}
	return result, nil
}

func (s *ProductService) GetByID(id int) (*dto.ProductResponse, error) {
	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, nil
	}
	images, err := s.imageRepo.GetByProductID(id)
	if err != nil {
		return nil, err
	}
	variants, err := s.variantRepo.GetByProductID(id)
	if err != nil {
		return nil, err
	}
	p.Images = images
	p.Variants = variants
	res := toProductResponse(*p)
	return &res, nil
}

func (s *ProductService) GetByCategoryID(categoryID int) ([]dto.ProductResponse, error) {
	products, err := s.repo.GetByCategoryID(categoryID)
	if err != nil {
		return nil, err
	}
	imagesMap, err := s.loadImagesForProducts(products)
	if err != nil {
		return nil, err
	}
	ids := make([]int, len(products))
	for i, p := range products {
		ids[i] = p.ID
	}
	variantsMap, err := s.variantRepo.GetByProductIDs(ids)
	if err != nil {
		return nil, err
	}
	result := make([]dto.ProductResponse, 0, len(products))
	for _, p := range products {
		p.Images = imagesMap[p.ID]
		p.Variants = variantsMap[p.ID]
		result = append(result, toProductResponse(p))
	}
	return result, nil
}

func (s *ProductService) loadImagesForProducts(products []domain.Product) (map[int][]domain.ProductImage, error) {
	ids := make([]int, len(products))
	for i, p := range products {
		ids[i] = p.ID
	}
	return s.imageRepo.GetByProductIDs(ids)
}

func (s *ProductService) Create(req dto.CreateProductRequest) (*dto.ProductResponse, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("field 'name' is required")
	}
	if req.SalePrice <= 0 {
		return nil, fmt.Errorf("field 'sale_price' must be greater than 0")
	}

	p, err := s.repo.Create(domain.Product{
		Name:        req.Name,
		Description: req.Description,
		SalePrice:   req.SalePrice,
		CategoryID:  req.CategoryID,
	})
	if err != nil {
		return nil, err
	}
	res := toProductResponse(*p)
	return &res, nil
}

func (s *ProductService) Update(id int, req dto.UpdateProductRequest) (*dto.ProductResponse, error) {
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
	description := existing.Description
	if req.Description != nil {
		description = req.Description
	}
	salePrice := existing.SalePrice
	if req.SalePrice != nil {
		salePrice = *req.SalePrice
	}
	categoryID := existing.CategoryID
	if req.CategoryID != nil {
		categoryID = req.CategoryID
	}
	active := existing.Active
	if req.Active != nil {
		active = *req.Active
	}

	updated, err := s.repo.Update(id, domain.Product{
		Name:        name,
		Description: description,
		SalePrice:   salePrice,
		CategoryID:  categoryID,
		Active:      active,
	})
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, nil
	}
	res := toProductResponse(*updated)
	return &res, nil
}

func (s *ProductService) Delete(id int) error {
	if s.cld == nil {
		log.Printf("[product.Delete] cloudinary client is nil – skipping image deletion for product %d", id)
	} else {
		images, err := s.imageRepo.GetByProductID(id)
		if err != nil {
			log.Printf("[product.Delete] could not fetch images for product %d: %v", id, err)
		} else {
			for _, img := range images {
				if img.PublicID == "" {
					continue
				}
				if err := s.cld.DeleteFile(context.Background(), img.PublicID); err != nil {
					log.Printf("[product.Delete] cloudinary delete failed for image id=%d public_id=%s: %v", img.ID, img.PublicID, err)
				}
			}
		}
	}
	_ = s.imageRepo.DeleteAllByProductID(id)
	return s.repo.SoftDelete(id)
}

// ──────────────────────────
// Images
// ──────────────────────────

func (s *ProductService) GetImages(productID int) ([]dto.ProductImageResponse, error) {
	existing, err := s.repo.GetByID(productID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("product not found")
	}
	images, err := s.imageRepo.GetByProductID(productID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.ProductImageResponse, 0, len(images))
	for _, img := range images {
		result = append(result, toProductImageResponse(img))
	}
	return result, nil
}

func (s *ProductService) AddImage(productID int, req dto.AddProductImageRequest) (*dto.ProductImageResponse, error) {
	if req.URL == "" {
		return nil, fmt.Errorf("field 'url' is required")
	}
	existing, err := s.repo.GetByID(productID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("product not found")
	}
	img, err := s.imageRepo.Add(productID, req.ColorID, req.URL, req.PublicID, req.SortOrder)
	if err != nil {
		return nil, err
	}
	res := toProductImageResponse(*img)
	return &res, nil
}

func (s *ProductService) UpdateImageColor(productID, imageID int, colorID *int) (*dto.ProductImageResponse, error) {
	img, err := s.imageRepo.UpdateColorID(productID, imageID, colorID)
	if err != nil {
		return nil, err
	}
	res := toProductImageResponse(*img)
	return &res, nil
}

func (s *ProductService) DeleteImage(productID, imageID int) error {
	if s.cld != nil {
		if img, err := s.imageRepo.GetOne(productID, imageID); err == nil && img != nil && img.PublicID != "" {
			_ = s.cld.DeleteFile(context.Background(), img.PublicID)
		}
	}
	return s.imageRepo.Delete(productID, imageID)
}

// ──────────────────────────
// Variants
// ──────────────────────────

func (s *ProductService) GetVariants(productID int) ([]dto.ProductVariantResponse, error) {
	existing, err := s.repo.GetByID(productID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("product not found")
	}
	variants, err := s.variantRepo.GetByProductID(productID)
	if err != nil {
		return nil, err
	}
	result := make([]dto.ProductVariantResponse, 0, len(variants))
	for _, v := range variants {
		result = append(result, toProductVariantResponse(v))
	}
	return result, nil
}

func (s *ProductService) AddVariant(productID int, req dto.CreateProductVariantRequest) (*dto.ProductVariantResponse, error) {
	existing, err := s.repo.GetByID(productID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, fmt.Errorf("product not found")
	}
	v, err := s.variantRepo.Create(productID, req.SizeID, req.ColorID, 0, req.PriceOverride)
	if err != nil {
		return nil, err
	}
	res := toProductVariantResponse(*v)
	return &res, nil
}

func (s *ProductService) UpdateVariant(productID, variantID int, req dto.UpdateProductVariantRequest) (*dto.ProductVariantResponse, error) {
	v, err := s.variantRepo.GetByID(variantID)
	if err != nil {
		return nil, err
	}
	if v == nil || v.ProductID != productID {
		return nil, fmt.Errorf("variant not found")
	}
	if req.PriceOverride != nil {
		if err := s.variantRepo.UpdatePriceOverride(variantID, req.PriceOverride); err != nil {
			return nil, err
		}
	}
	updated, err := s.variantRepo.GetByID(variantID)
	if err != nil {
		return nil, err
	}
	res := toProductVariantResponse(*updated)
	return &res, nil
}

func (s *ProductService) DeleteVariant(productID, variantID int) error {
	v, err := s.variantRepo.GetByID(variantID)
	if err != nil {
		return err
	}
	if v == nil || v.ProductID != productID {
		return fmt.Errorf("variant not found")
	}
	hasMovements, err := s.inventoryRepo.HasMovements(variantID)
	if err != nil {
		return err
	}
	if hasMovements {
		return fmt.Errorf("has_movements: la variante tiene movimientos de inventario y no puede eliminarse")
	}
	return s.variantRepo.Delete(variantID)
}

// ──────────────────────────
// Mappers
// ──────────────────────────

func toProductResponse(p domain.Product) dto.ProductResponse {
	imageResponses := make([]dto.ProductImageResponse, 0, len(p.Images))
	for _, img := range p.Images {
		imageResponses = append(imageResponses, toProductImageResponse(img))
	}
	variantResponses := make([]dto.ProductVariantResponse, 0, len(p.Variants))
	for _, v := range p.Variants {
		variantResponses = append(variantResponses, toProductVariantResponse(v))
	}
	return dto.ProductResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		SalePrice:   p.SalePrice,
		CategoryID:  p.CategoryID,
		Active:      p.Active,
		Images:      imageResponses,
		Variants:    variantResponses,
	}
}

func toProductImageResponse(img domain.ProductImage) dto.ProductImageResponse {
	return dto.ProductImageResponse{
		ID:        img.ID,
		ProductID: img.ProductID,
		ColorID:   img.ColorID,
		URL:       img.URL,
		PublicID:  img.PublicID,
		SortOrder: img.SortOrder,
	}
}

func toProductVariantResponse(v domain.ProductVariant) dto.ProductVariantResponse {
	return dto.ProductVariantResponse{
		ID:            v.ID,
		ProductID:     v.ProductID,
		SizeID:        v.SizeID,
		SizeName:      v.SizeName,
		ColorID:       v.ColorID,
		ColorName:     v.ColorName,
		ColorHex:      v.ColorHex,
		Stock:         v.Stock,
		PriceOverride: v.PriceOverride,
	}
}

func (s *ProductService) GetActiveWithStock() ([]dto.ProductResponse, error) {
    products, err := s.repo.GetActiveWithStock()
    if err != nil {
        return nil, err
    }
    imagesMap, err := s.loadImagesForProducts(products)
    if err != nil {
        return nil, err
    }
    ids := make([]int, len(products))
    for i, p := range products {
        ids[i] = p.ID
    }
    variantsMap, err := s.variantRepo.GetByProductIDs(ids)
    if err != nil {
        return nil, err
    }
    result := make([]dto.ProductResponse, 0, len(products))
    for _, p := range products {
        p.Images = imagesMap[p.ID]
        p.Variants = variantsMap[p.ID]
        result = append(result, toProductResponse(p))
    }
    return result, nil
}

package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/brayanzuritadev/StockManager/internal/handlers/helpers"
	"github.com/brayanzuritadev/StockManager/internal/models/dto"
	"github.com/brayanzuritadev/StockManager/internal/repositories"
	"github.com/brayanzuritadev/StockManager/internal/services"
	"github.com/brayanzuritadev/StockManager/pkg/cloudinarypkg"
)

// ProductHandler routes all /products/* requests.
//
//	GET    /products               ?active=true
//	POST   /products
//	GET    /products/{id}
//	PUT    /products/{id}
//	DELETE /products/{id}
//	GET    /products/category/{category_id}
//	GET    /products/{id}/images
//	POST   /products/{id}/images
//	DELETE /products/{id}/images/{image_id}
//	GET    /products/{id}/variants
//	POST   /products/{id}/variants
//	PUT    /products/{id}/variants/{variant_id}
//	DELETE /products/{id}/variants/{variant_id}
func ProductHandler(ctx context.Context, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	db := ctx.Value(dto.KeyDatabase).(*sql.DB)
	method := ctx.Value(dto.KeyMethod).(string)
	path := ctx.Value(dto.KeyPath).(string)

	svc := newProductService(ctx, db)

	remaining := strings.Trim(strings.TrimPrefix(path, "products"), "/")
	
	segments := []string{}
	if remaining != "" {
		segments = strings.Split(remaining, "/")
	}

	if method == http.MethodOptions {
		return helpers.PreflightResponse(req)
	}

	// ── GET /products | POST /products ────────────────────────────────────
	if len(segments) == 0 {
		switch method {
		case http.MethodGet:
			onlyActive := req.QueryStringParameters["active"] == "true"
			onlyWithStock := req.QueryStringParameters["with_stock"] == "true"
			if onlyWithStock && onlyActive {
				products, err := svc.GetActiveWithStock()
				if err != nil {
					return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
				}
				return helpers.JsonResponse(200, products, req)
			}
			products, err := svc.GetAll(onlyActive)
			if err != nil {
				return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(200, products, req)

		case http.MethodPost:
			var body dto.CreateProductRequest
			if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
				return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])
			}
			product, err := svc.Create(body)
			if err != nil {
				return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(201, product, req)
		}
		return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
	}

	// ── GET /products/category/{category_id} ──────────────────────────────
	if len(segments) == 2 && segments[0] == "category" {
		if method != http.MethodGet {
			return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
		}
		categoryID, err := strconv.Atoi(segments[1])
		if err != nil {
			return helpers.ErrorResponse(400, fmt.Sprintf("invalid category id: %s", segments[1]), req.Headers["origin"])
		}
		products, err := svc.GetByCategoryID(categoryID)
		if err != nil {
			return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
		}
		return helpers.JsonResponse(200, products, req)
	}

	// ── /products/{id} ────────────────────────────────────────────────────
	if len(segments) == 1 {
		id, err := strconv.Atoi(segments[0])
		if err != nil {
			return helpers.ErrorResponse(400, fmt.Sprintf("invalid product id: %s", segments[0]), req.Headers["origin"])
		}

		switch method {
		case http.MethodGet:
			product, err := svc.GetByID(id)
			if err != nil {
				return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
			}
			if product == nil {
				return helpers.ErrorResponse(404, "product not found", req.Headers["origin"])
			}
			return helpers.JsonResponse(200, product, req)

		case http.MethodPut:
			var body dto.UpdateProductRequest
			if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
				return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])

			}
			product, err := svc.Update(id, body)
			if err != nil {
				return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
			}
			if product == nil {
				return helpers.ErrorResponse(404, "product not found", req.Headers["origin"])
			}
			return helpers.JsonResponse(200, product, req)

		case http.MethodDelete:
			if err := svc.Delete(id); err != nil {
				return helpers.ErrorResponse(404, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(200, map[string]string{"message": "product deleted"}, req)
		}
		return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
	}

	// ── /products/{id}/images ─────────────────────────────────────────────
	if len(segments) == 2 && segments[1] == "images" {
		id, err := strconv.Atoi(segments[0])
		if err != nil {
			return helpers.ErrorResponse(400, fmt.Sprintf("invalid product id: %s", segments[0]), req.Headers["origin"])
		}
		switch method {
		case http.MethodGet:
			images, err := svc.GetImages(id)
			if err != nil {
				return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(200, images, req)
		case http.MethodPost:
			var body dto.AddProductImageRequest
			if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
				return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])
			}
			img, err := svc.AddImage(id, body)
			if err != nil {
				return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(201, img, req)
		}
		return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
	}

	// ── /products/{id}/images/{image_id} ──────────────────────────────────
	if len(segments) == 3 && segments[1] == "images" {
		productID, err := strconv.Atoi(segments[0])
		if err != nil {
			return helpers.ErrorResponse(400, "invalid product id", req.Headers["origin"])
		}
		imageID, err := strconv.Atoi(segments[2])
		if err != nil {
			return helpers.ErrorResponse(400, "invalid image id", req.Headers["origin"])
		}
		switch method {
		case http.MethodDelete:
			if err := svc.DeleteImage(productID, imageID); err != nil {
				return helpers.ErrorResponse(404, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(200, map[string]string{"message": "image deleted"}, req)
		case http.MethodPatch:
			var body struct {
				ColorID *int `json:"color_id"`
			}
			if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
				return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])
			}
			img, err := svc.UpdateImageColor(productID, imageID, body.ColorID)
			if err != nil {
				return helpers.ErrorResponse(404, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(200, img, req)
		}
		return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
	}

	// ── /products/{id}/variants ───────────────────────────────────────────
	if len(segments) == 2 && segments[1] == "variants" {
		id, err := strconv.Atoi(segments[0])
		if err != nil {
			return helpers.ErrorResponse(400, fmt.Sprintf("invalid product id: %s", segments[0]), req.Headers["origin"])
		}
		switch method {
		case http.MethodGet:
			variants, err := svc.GetVariants(id)
			if err != nil {
				return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(200, variants, req)
		case http.MethodPost:
			var body dto.CreateProductVariantRequest
			if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
				return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])
			}
			v, err := svc.AddVariant(id, body)
			if err != nil {
				return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(201, v, req)
		}
		return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
	}

	// ── /products/{id}/variants/{variant_id} ──────────────────────────────
	if len(segments) == 3 && segments[1] == "variants" {
		productID, err := strconv.Atoi(segments[0])
		if err != nil {
			return helpers.ErrorResponse(400, "invalid product id", req.Headers["origin"])
		}
		variantID, err := strconv.Atoi(segments[2])
		if err != nil {
			return helpers.ErrorResponse(400, "invalid variant id", req.Headers["origin"])
		}
		switch method {
		case http.MethodPut:
			var body dto.UpdateProductVariantRequest
			if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
				return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])
			}
			v, err := svc.UpdateVariant(productID, variantID, body)
			if err != nil {
				return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(200, v, req)
		case http.MethodDelete:
			if err := svc.DeleteVariant(productID, variantID); err != nil {
				if strings.HasPrefix(err.Error(), "has_movements:") {
					return helpers.ErrorResponse(409, "La variante tiene movimientos de inventario y no puede eliminarse", req.Headers["origin"])

				}
				return helpers.ErrorResponse(404, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(200, map[string]string{"message": "variant deleted"}, req)
		}
		return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
	}

	return helpers.ErrorResponse(404, "route not found", req.Headers["origin"])
}

func newProductService(ctx context.Context, db *sql.DB) services.IProductService {
	cloudName, _ := ctx.Value(dto.KeyCloudName).(string)
	apiKey, _ := ctx.Value(dto.KeyCloudAPIKey).(string)
	apiSecret, _ := ctx.Value(dto.KeyCloudAPISecret).(string)
	var cld services.ICloudinaryDeleter
	if c, err := cloudinarypkg.NewClient(cloudName, apiKey, apiSecret); err == nil {
		cld = c
	}
	return services.NewProductService(
		repositories.NewProductRepository(db),
		repositories.NewProductImageRepository(db),
		repositories.NewProductVariantRepository(db),
		repositories.NewInventoryMovementRepository(db),
		cld,
	)
}

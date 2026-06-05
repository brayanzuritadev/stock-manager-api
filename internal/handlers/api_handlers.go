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
)

// ── Sale Handler ──────────────────────────────────────────────────────────
//
//	GET    /sales
//	POST   /sales
//	GET    /sales/{id}
//	PATCH  /sales/{id}/status
func SaleHandler(ctx context.Context, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	db := ctx.Value(dto.KeyDatabase).(*sql.DB)
	method := ctx.Value(dto.KeyMethod).(string)
	path := ctx.Value(dto.KeyPath).(string)

	svc := services.NewSaleService(
		repositories.NewSaleRepository(db),
		repositories.NewProductVariantRepository(db),
		repositories.NewInventoryMovementRepository(db),
	)

	remaining := strings.Trim(strings.TrimPrefix(path, "sales"), "/")
	segments := []string{}
	if remaining != "" {
		segments = strings.Split(remaining, "/")
	}

	if method == http.MethodOptions {
		return helpers.PreflightResponse(req)
	}

	if len(segments) == 0 {
		switch method {
		case http.MethodGet:
			sales, err := svc.GetAll()
			if err != nil {
				return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(200, sales, req)

		case http.MethodPost:
			var body dto.CreateSaleRequest
			if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
				return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])
			}
			var items []services.SaleItemInput
			for _, it := range body.Items {
				items = append(items, services.SaleItemInput{
					ProductVariantID: it.ProductVariantID,
					Quantity:         it.Quantity,
					UnitPrice:        it.UnitPrice,
					Discount:         it.Discount,
				})
			}
			sale, err := svc.Create(body.UserID, body.Channel, body.Status, body.TotalDiscount, body.Notes, items)
			if err != nil {
				return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(201, sale, req)
		}
		return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
	}

	id, err := strconv.Atoi(segments[0])
	if err != nil {
		return helpers.ErrorResponse(400, fmt.Sprintf("invalid sale id: %s", segments[0]), req.Headers["origin"])
	}

	// PATCH /sales/{id}/status
	if len(segments) == 2 && segments[1] == "status" && method == http.MethodPatch {
		var body struct {
			Status string `json:"status"`
		}
		if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
			return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])
		}
		if err := svc.UpdateStatus(id, body.Status); err != nil {
			return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
		}
		return helpers.JsonResponse(200, map[string]string{"message": "status updated"}, req)
	}

	switch method {
	case http.MethodGet:
		sale, err := svc.GetByID(id)
		if err != nil {
			return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
		}
		if sale == nil {
			return helpers.ErrorResponse(404, "sale not found", req.Headers["origin"])
		}
		return helpers.JsonResponse(200, sale, req)
	case http.MethodDelete:
		if err := svc.Delete(id); err != nil {
			return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
		}
		return helpers.JsonResponse(200, map[string]string{"message": "sale deleted"}, req)					
	}
	return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
}

// ── Inventory Movement Handler ────────────────────────────────────────────
//
//	GET    /inventory/movements
//	POST   /inventory/movements
//	GET    /inventory/movements/{id}
func InventoryMovementHandler(ctx context.Context, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	db := ctx.Value(dto.KeyDatabase).(*sql.DB)
	method := ctx.Value(dto.KeyMethod).(string)
	path := ctx.Value(dto.KeyPath).(string)

	svc := services.NewInventoryMovementService(
		repositories.NewInventoryMovementRepository(db),
		repositories.NewProductVariantRepository(db),
	)

	remaining := strings.Trim(strings.TrimPrefix(path, "inventory/movements"), "/")
	segments := []string{}
	if remaining != "" {
		segments = strings.Split(remaining, "/")
	}

	if method == http.MethodOptions {
		return helpers.PreflightResponse(req)
	}

	if len(segments) == 0 {
		switch method {
		case http.MethodGet:
			list, err := svc.GetAll()
			if err != nil {
				return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(200, list, req)

		case http.MethodPost:
			var body dto.CreateInventoryMovementRequest
			if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
				return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])
			}
			mv, err := svc.Create(body.ProductVariantID, body.Type, body.Quantity, body.UnitCost, body.ReferenceType, body.ReferenceID)
			if err != nil {
				return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(201, mv, req)
		}
		return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
	}

	id, err := strconv.Atoi(segments[0])
	if err != nil {
		return helpers.ErrorResponse(400, fmt.Sprintf("invalid movement id: %s", segments[0]), req.Headers["origin"])
	}

	if method == http.MethodGet {
		mv, err := svc.GetByID(id)
		if err != nil {
			return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])	
		}
		if mv == nil {
			return helpers.ErrorResponse(404, "inventory movement not found", req.Headers["origin"])
		}
		return helpers.JsonResponse(200, mv, req)
	}
	return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
}

// ── Size Handler ──────────────────────────────────────────────────────────
//
//	GET    /sizes
//	POST   /sizes
//	DELETE /sizes/{id}
func SizeHandler(ctx context.Context, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	db := ctx.Value(dto.KeyDatabase).(*sql.DB)
	method := ctx.Value(dto.KeyMethod).(string)
	path := ctx.Value(dto.KeyPath).(string)

	repo := repositories.NewSizeRepository(db)

	remaining := strings.Trim(strings.TrimPrefix(path, "sizes"), "/")
	segments := []string{}
	if remaining != "" {
		segments = strings.Split(remaining, "/")
	}

	if method == http.MethodOptions {
		return helpers.PreflightResponse(req)
	}

	if len(segments) == 0 {
		switch method {
		case http.MethodGet:
			list, err := repo.GetAll()
			if err != nil {
				return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(200, list, req)

		case http.MethodPost:
			var body dto.CreateSizeRequest
			if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
				return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])
			}
			sz, err := repo.Create(body.Name, body.SortOrder)
			if err != nil {
				return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(201, sz, req)
		}
		return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
	}

	id, err := strconv.Atoi(segments[0])
	if err != nil {
		return helpers.ErrorResponse(400, fmt.Sprintf("invalid size id: %s", segments[0]), req.Headers["origin"])
	}

	if method == http.MethodDelete {
		if err := repo.Delete(id); err != nil {
			return helpers.ErrorResponse(404, err.Error(), req.Headers["origin"])
		}
		return helpers.JsonResponse(200, map[string]string{"message": "size deleted"}, req)
	}

	if method == http.MethodPut {
		var body dto.CreateSizeRequest
		if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
			return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])
		}
		sz, err := repo.Update(id, body.Name, body.SortOrder)
		if err != nil {
			return helpers.ErrorResponse(404, err.Error(), req.Headers["origin"])
		}
		return helpers.JsonResponse(200, sz, req)
	}

	return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
}

// ── Color Handler ─────────────────────────────────────────────────────────
//
//	GET    /colors
//	POST   /colors
//	DELETE /colors/{id}
func ColorHandler(ctx context.Context, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	db := ctx.Value(dto.KeyDatabase).(*sql.DB)
	method := ctx.Value(dto.KeyMethod).(string)
	path := ctx.Value(dto.KeyPath).(string)

	repo := repositories.NewColorRepository(db)

	remaining := strings.Trim(strings.TrimPrefix(path, "colors"), "/")
	segments := []string{}
	if remaining != "" {
		segments = strings.Split(remaining, "/")
	}

	if method == http.MethodOptions {
		return helpers.PreflightResponse(req)
	}

	if len(segments) == 0 {
		switch method {
		case http.MethodGet:
			list, err := repo.GetAll()
			if err != nil {
				return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(200, list, req)

		case http.MethodPost:
			var body dto.CreateColorRequest
			if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
				return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])
			}
			c, err := repo.Create(body.Name, body.HexCode)
			if err != nil {
				return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(201, c, req)
		}
		return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
	}

	id, err := strconv.Atoi(segments[0])
	if err != nil {
		return helpers.ErrorResponse(400, fmt.Sprintf("invalid color id: %s", segments[0]), req.Headers["origin"])
	}

	if method == http.MethodDelete {
		if err := repo.Delete(id); err != nil {
			return helpers.ErrorResponse(404, err.Error(), req.Headers["origin"])
		}
		return helpers.JsonResponse(200, map[string]string{"message": "color deleted"}, req)
	}

	if method == http.MethodPut {
		var body dto.CreateColorRequest
		if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
			return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])
		}
		c, err := repo.Update(id, body.Name, body.HexCode)
		if err != nil {
			return helpers.ErrorResponse(404, err.Error(), req.Headers["origin"])
		}
		return helpers.JsonResponse(200, c, req)
	}

	return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
}

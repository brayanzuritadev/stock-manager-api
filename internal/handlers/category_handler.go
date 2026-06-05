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

// CategoryHandler routes all /categories/* requests.
//
//	GET    /categories
//	POST   /categories
//	GET    /categories/{id}
//	PUT    /categories/{id}
//	DELETE /categories/{id}
func CategoryHandler(ctx context.Context, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	db := ctx.Value(dto.KeyDatabase).(*sql.DB)
	method := ctx.Value(dto.KeyMethod).(string)
	path := ctx.Value(dto.KeyPath).(string)

	svc := newCategoryService(db)

	remaining := strings.Trim(strings.TrimPrefix(path, "categories"), "/")
	segments := []string{}
	if remaining != "" {
		segments = strings.Split(remaining, "/")
	}

	if method == http.MethodOptions {
		return helpers.PreflightResponse(req)
	}

	// ── GET /categories | POST /categories ────────────────────────────────
	if method == http.MethodOptions {
		return helpers.PreflightResponse(req)
	}
	if len(segments) == 0 {
		switch method {
		case http.MethodGet:
			categories, err := svc.GetAll()
			if err != nil {
				return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(200, categories, req)

		case http.MethodPost:
			var body dto.CreateCategoryRequest
			if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
				return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])
			}
			category, err := svc.Create(body)
			if err != nil {
				return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(201, category, req)
		}
		return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
	}

	// ── /categories/{id} ──────────────────────────────────────────────────
	if len(segments) == 1 {
		id, err := strconv.Atoi(segments[0])
		if err != nil {
			return helpers.ErrorResponse(400, fmt.Sprintf("invalid category id: %s", segments[0]), req.Headers["origin"])
		}

		switch method {
		case http.MethodGet:
			category, err := svc.GetByID(id)
			if err != nil {
				return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
			}
			if category == nil {
				return helpers.ErrorResponse(404, "category not found", req.Headers["origin"])
			}
			return helpers.JsonResponse(200, category, req)

		case http.MethodPut:
			var body dto.UpdateCategoryRequest
			if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
				return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])
			}
			category, err := svc.Update(id, body)
			if err != nil {
				return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
			}
			if category == nil {
				return helpers.ErrorResponse(404, "category not found", req.Headers["origin"])
			}
			return helpers.JsonResponse(200, category, req)

		case http.MethodDelete:
			if err := svc.Delete(id); err != nil {
				if err == sql.ErrNoRows {
					return helpers.ErrorResponse(404, "category not found", req.Headers["origin"])
				}
				return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(204, nil, req)
		}
		return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
	}

	return helpers.ErrorResponse(404, "route not found", req.Headers["origin"])
}

func newCategoryService(db *sql.DB) services.ICategoryService {
	return services.NewCategoryService(repositories.NewCategoryRepository(db))
}

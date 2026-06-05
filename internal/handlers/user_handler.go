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

// UserHandler routes all /users/* requests.
//
//	GET    /users
//	POST   /users
//	GET    /users/{id}
//	PUT    /users/{id}
//	DELETE /users/{id}
func UserHandler(ctx context.Context, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	db := ctx.Value(dto.KeyDatabase).(*sql.DB)
	method := ctx.Value(dto.KeyMethod).(string)
	path := ctx.Value(dto.KeyPath).(string)

	svc := newUserService(db)

	remaining := strings.Trim(strings.TrimPrefix(path, "users"), "/")
	segments := []string{}
	if remaining != "" {
		segments = strings.Split(remaining, "/")
	}

	if method == http.MethodOptions {
		return helpers.PreflightResponse(req)
	}

	// ── GET /users | POST /users ──────────────────────────────────────────
	if method == http.MethodOptions {
		return helpers.PreflightResponse(req)
	}
	if len(segments) == 0 {
		switch method {
		case http.MethodGet:
			users, err := svc.GetAll()
			if err != nil {
				return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(200, users, req)

		case http.MethodPost:
			var body dto.CreateUserRequest
			if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
				return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])
			}
			user, err := svc.Create(body)
			if err != nil {
				return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(201, user, req)
		}
		return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
	}

	// ── /users/{id} ───────────────────────────────────────────────────────
	if len(segments) == 1 {
		id, err := strconv.Atoi(segments[0])
		if err != nil {
			return helpers.ErrorResponse(400, fmt.Sprintf("invalid user id: %s", segments[0]), req.Headers["origin"])
		}

		switch method {
		case http.MethodGet:
			user, err := svc.GetByID(id)
			if err != nil {
				return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
			}
			if user == nil {
				return helpers.ErrorResponse(404, "user not found", req.Headers["origin"])
			}
			return helpers.JsonResponse(200, user, req)

		case http.MethodPut:
			var body dto.UpdateUserRequest
			if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
				return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])
			}
			user, err := svc.Update(id, body)
			if err != nil {
				return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
			}
			if user == nil {
				return helpers.ErrorResponse(404, "user not found", req.Headers["origin"])
			}
			return helpers.JsonResponse(200, user, req)

		case http.MethodDelete:
			if err := svc.Delete(id); err != nil {
				return helpers.ErrorResponse(404, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(200, map[string]string{"message": "user deleted"}, req)
		}
		return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
	}

	return helpers.ErrorResponse(404, "route not found", req.Headers["origin"])
}

func newUserService(db *sql.DB) services.IUserService {
	return services.NewUserService(repositories.NewUserRepository(db))
}

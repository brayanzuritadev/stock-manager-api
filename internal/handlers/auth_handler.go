package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/brayanzuritadev/StockManager/internal/models/dto"
	"github.com/brayanzuritadev/StockManager/internal/repositories"
	"github.com/brayanzuritadev/StockManager/internal/services"
	"github.com/brayanzuritadev/StockManager/internal/handlers/helpers"
)

// AuthHandler routes all /auth/* requests.
//
//	POST /auth/login
//	POST /auth/register
func AuthHandler(ctx context.Context, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	db := ctx.Value(dto.KeyDatabase).(*sql.DB)
	method := ctx.Value(dto.KeyMethod).(string)
	path := ctx.Value(dto.KeyPath).(string)
	jwtSecret := ctx.Value(dto.KeyJWTSign).(string)

	svc := newAuthService(db)

	remaining := strings.Trim(strings.TrimPrefix(path, "auth"), "/")

	if method == http.MethodOptions {
		return helpers.PreflightResponse(req)
	}

	switch remaining {
	case "login":
		var body dto.LoginRequest
		if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
			return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])
		}
		resp, err := svc.Login(body, jwtSecret)
		if err != nil {
			return helpers.ErrorResponse(401, err.Error(), req.Headers["origin"])
		}
		return helpers.JsonResponse(200, resp, req)

	/*case "register":
		var body dto.RegisterRequest
		if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
			return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])
		}
		resp, err := svc.Register(body, jwtSecret)
		if err != nil {
			return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
		}
		return helpers.JsonResponse(201, resp, req)*/

	default:
		return helpers.ErrorResponse(404, "route not found", req.Headers["origin"])
	}
}

func newAuthService(db *sql.DB) services.IAuthService {
	return services.NewAuthService(repositories.NewUserRepository(db))
}

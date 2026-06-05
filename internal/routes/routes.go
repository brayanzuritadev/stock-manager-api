package routes

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/brayanzuritadev/StockManager/internal/handlers"
)

const (
	authRoutePrefix     = "auth"
	userRoutePrefix     = "users"
	productRoutePrefix  = "products"
	categoryRoutePrefix = "categories"
	cartRoutePrefix     = "carts"
	saleRoutePrefix     = "sales"
	inventoryMovePrefix = "inventory/movements"
	sizeRoutePrefix     = "sizes"
	colorRoutePrefix    = "colors"
	reportRoutePrefix   = "reports"
)

func RouteRequest(path string, ctx context.Context, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	fmt.Printf("Routing request: %s %s\n", req.HTTPMethod, path)
	switch {
	case strings.HasPrefix(path, authRoutePrefix):
		return handlers.AuthHandler(ctx, req)
	case strings.HasPrefix(path, userRoutePrefix):
		return handlers.UserHandler(ctx, req)
	case strings.HasPrefix(path, productRoutePrefix):
		return handlers.ProductHandler(ctx, req)
	case strings.HasPrefix(path, categoryRoutePrefix):
		return handlers.CategoryHandler(ctx, req)
	case strings.HasPrefix(path, cartRoutePrefix):
		return handlers.CartHandler(ctx, req)
	case strings.HasPrefix(path, saleRoutePrefix):
		return handlers.SaleHandler(ctx, req)
	case strings.HasPrefix(path, inventoryMovePrefix):
		return handlers.InventoryMovementHandler(ctx, req)
	case strings.HasPrefix(path, sizeRoutePrefix):
		return handlers.SizeHandler(ctx, req)
	case strings.HasPrefix(path, colorRoutePrefix):
		return handlers.ColorHandler(ctx, req)
	case strings.HasPrefix(path, reportRoutePrefix):
		return handlers.ReportHandler(ctx, req)
	default:
		return &events.APIGatewayProxyResponse{StatusCode: 404, Body: "Route not found"}, nil
	}
}

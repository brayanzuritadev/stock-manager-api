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

// allowedChannels for sale confirmation
var allowedChannels = map[string]bool{"web": true, "whatsapp": true, "tienda": true}

func newCartService(db *sql.DB) services.ICartService {
	return services.NewCartService(repositories.NewCartRepository(db))
}

// CartHandler routes all /carts/* requests.
//
// Supported routes:
//   GET    carts
//   POST   carts
//   GET    carts/{id}
//   PUT    carts/{id}
//   DELETE carts/{id}
//   POST   carts/{id}/share
//   GET    carts/shared/{link}
//   GET    carts/{cart_id}/items
//   POST   carts/{cart_id}/items
//   PUT    carts/{cart_id}/items/{item_id}
//   DELETE carts/{cart_id}/items/{item_id}

func CartHandler(ctx context.Context, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	db := ctx.Value(dto.KeyDatabase).(*sql.DB)
	method := ctx.Value(dto.KeyMethod).(string)
	path := ctx.Value(dto.KeyPath).(string)

	svc := newCartService(db)

	// Strip leading prefix "carts" to get the rest of the path segments
	remaining := strings.TrimPrefix(path, "carts")
	remaining = strings.Trim(remaining, "/")
	segments := []string{}
	if remaining != "" {
		segments = strings.Split(remaining, "/")
	}

	if method == http.MethodOptions {
		return helpers.PreflightResponse(req)
	}

	// ── GET /carts / POST /carts ──────────────────────────────────────────
	if len(segments) == 0 {
		switch method {
		case http.MethodGet:
			carts, err := svc.GetAll()
			if err != nil {
				return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(200, carts, req)

		case http.MethodPost:
			var body struct {
				UserID *int `json:"user_id"`
			}
			_ = json.Unmarshal([]byte(req.Body), &body)
			cart, err := svc.Create(body.UserID)
			if err != nil {
				return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(201, cart, req)
		}
		return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
	}

	// ── GET /carts/shared/{link}  PATCH /carts/shared/{link} ─────────────
	if segments[0] == "shared" && len(segments) == 2 {
		switch method {
		case http.MethodGet:
			cart, err := svc.GetBySharedLink(segments[1])
			if err != nil {
				return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
			}
			if cart == nil {
				return helpers.ErrorResponse(404, "cart not found", req.Headers["origin"])
			}
			return helpers.JsonResponse(200, cart, req)

		case http.MethodPatch:
			var body struct {
				Notes string `json:"notes"`
			}
			if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
				return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])
			}
			cart, err := svc.UpdateNotesByLink(segments[1], body.Notes)
			if err != nil {
				return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
			}
			if cart == nil {
				return helpers.ErrorResponse(404, "cart not found or already confirmed", req.Headers["origin"])
			}
			return helpers.JsonResponse(200, cart, req)
		}
		return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
	}

	// All remaining routes start with a numeric cart ID
	cartID, err := strconv.Atoi(segments[0])
	if err != nil {
		return helpers.ErrorResponse(400, fmt.Sprintf("invalid cart id: %s", segments[0]), req.Headers["origin"])
	}

	// ── /carts/{id} ───────────────────────────────────────────────────────
	if len(segments) == 1 {
		switch method {
		case http.MethodGet:
			cart, err := svc.GetByID(cartID)
			if err != nil {
				return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
			}
			if cart == nil {
				return helpers.ErrorResponse(404, "cart not found", req.Headers["origin"])
			}
			return helpers.JsonResponse(200, cart, req)

		case http.MethodPut:
			var body struct {
				Status string `json:"status"`
			}
			if err := json.Unmarshal([]byte(req.Body), &body); err != nil || body.Status == "" {
				return helpers.ErrorResponse(400, "field 'status' is required", req.Headers["origin"])
			}
			cart, err := svc.UpdateStatus(cartID, body.Status)
			if err != nil {
				return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
			}
			if cart == nil {
				return helpers.ErrorResponse(404, "cart not found", req.Headers["origin"])
			}
			return helpers.JsonResponse(200, cart, req)

		case http.MethodDelete:
			if err := svc.Delete(cartID); err != nil {
				return helpers.ErrorResponse(404, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(200, map[string]string{"message": "cart deleted"}, req)
		}
		return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
	}

	// ── /carts/{id}/share ────────────────────────────────────────────────
	if len(segments) == 2 && segments[1] == "share" {
		if method != http.MethodPost {
			return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
		}
		cart, err := svc.GenerateSharedLink(cartID)
		if err != nil {
			return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
		}
		return helpers.JsonResponse(200, map[string]interface{}{
			"cart_id":     cart.ID,
			"shared_link": cart.SharedLink,
		}, req)
	}

	// ── /carts/{cart_id}/items ────────────────────────────────────────────
	if len(segments) == 2 && segments[1] == "items" {
		switch method {
		case http.MethodGet:
			items, err := svc.GetItems(cartID)
			if err != nil {
				return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(200, items, req)

		case http.MethodPost:
			var body struct {
				ProductVariantID int     `json:"product_variant_id"`
				Quantity         int     `json:"quantity"`
				UnitPrice        float64 `json:"unit_price"`
				Discount         float64 `json:"discount"`
			}
			if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
				return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])

			}
			if body.ProductVariantID == 0 {
				return helpers.ErrorResponse(400, "field 'product_variant_id' is required", req.Headers["origin"])
			}
			item, err := svc.AddItem(cartID, body.ProductVariantID, body.Quantity, body.UnitPrice, body.Discount)
			if err != nil {
				return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(201, item, req)
		}
		return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
	}

	// ── /carts/{cart_id}/items/{item_id} ──────────────────────────────────
	if len(segments) == 3 && segments[1] == "items" {
		itemID, err := strconv.Atoi(segments[2])
		if err != nil {
			return helpers.ErrorResponse(400, "invalid item id", req.Headers["origin"])
		}

		switch method {
		case http.MethodPut:
			var body struct {
				Quantity int     `json:"quantity"`
				Discount float64 `json:"discount"`
			}
			if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
				return helpers.ErrorResponse(400, "invalid request body", req.Headers["origin"])
			}
			item, err := svc.UpdateItem(cartID, itemID, body.Quantity, body.Discount)
			if err != nil {
				return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(200, item, req)

		case http.MethodDelete:
			if err := svc.DeleteItem(cartID, itemID); err != nil {
				return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
			}
			return helpers.JsonResponse(200, map[string]string{"message": "item deleted"}, req)
		}
		return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
	}

	// ── POST /carts/{id}/confirm → create sale, mark cart completed ───────
	if len(segments) == 2 && segments[1] == "confirm" {
		if method != http.MethodPost {
			return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
		}

		var body struct {
			Channel string  `json:"channel"`
			Notes   *string `json:"notes"`
		}
		_ = json.Unmarshal([]byte(req.Body), &body)
		if body.Channel == "" {
			body.Channel = "whatsapp"
		}
		if !allowedChannels[body.Channel] {
			return helpers.ErrorResponse(400, "channel must be one of: web, whatsapp, tienda", req.Headers["origin"])
		}

		// Load cart + items
		cart, err := svc.GetByID(cartID)
		if err != nil {
			return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
		}
		if cart == nil {
			return helpers.ErrorResponse(404, "cart not found", req.Headers["origin"])
		}
		if cart.Status != "pending" {
			return helpers.ErrorResponse(409, fmt.Sprintf("cart is already %s", cart.Status), req.Headers["origin"])
		}
		if len(cart.Items) == 0 {
			return helpers.ErrorResponse(400, "cart has no items", req.Headers["origin"])
		}

		// Build sale items
		saleItems := make([]services.SaleItemInput, len(cart.Items))
		for i, it := range cart.Items {
			saleItems[i] = services.SaleItemInput{
				ProductVariantID: it.ProductVariantID,
				Quantity:         it.Quantity,
				UnitPrice:        it.UnitPrice,
				Discount:         it.Discount,
			}
		}

		// Create sale (validates stock, records movements, deducts stock)
		saleSvc := services.NewSaleService(
			repositories.NewSaleRepository(db),
			repositories.NewProductVariantRepository(db),
			repositories.NewInventoryMovementRepository(db),
		)
		// If admin didn't provide a sale note, fall back to the cart's customer note
		noteToUse := body.Notes
		if noteToUse == nil {
			noteToUse = cart.Notes
		}
		ch := body.Channel
		sale, err := saleSvc.Create(cart.UserID, &ch, nil, 0, noteToUse, saleItems)
		if err != nil {
			return helpers.ErrorResponse(400, err.Error(), req.Headers["origin"])
		}

		// Mark cart as completed
		if _, err := svc.UpdateStatus(cartID, "completed"); err != nil {
			return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
		}

		return helpers.JsonResponse(201, sale, req)
	}

	return helpers.ErrorResponse(404, "route not found", req.Headers["origin"])
}

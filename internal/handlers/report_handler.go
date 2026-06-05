package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/brayanzuritadev/StockManager/internal/handlers/helpers"
	"github.com/brayanzuritadev/StockManager/internal/models/dto"
)

// ──────────────────────────────────────────────────────────────────────────────
// Response types
// ──────────────────────────────────────────────────────────────────────────────

type salesReport struct {
	PeriodFrom    string       `json:"period_from"`
	PeriodTo      string       `json:"period_to"`
	TotalRevenue  float64      `json:"total_revenue"`
	TotalSales    int          `json:"total_sales"`
	AvgOrder      float64      `json:"avg_order"`
	TotalDiscount float64      `json:"total_discount"`
	ByChannel     []channelRow `json:"by_channel"`
	ByStatus      []statusRow  `json:"by_status"`
	ByDay         []dayRow     `json:"by_day"`
	TopProducts   []topProduct `json:"top_products"`
}

type channelRow struct {
	Channel string  `json:"channel"`
	Count   int     `json:"count"`
	Revenue float64 `json:"revenue"`
}

type statusRow struct {
	Status  string  `json:"status"`
	Count   int     `json:"count"`
	Revenue float64 `json:"revenue"`
}

type dayRow struct {
	Day     string  `json:"day"`
	Count   int     `json:"count"`
	Revenue float64 `json:"revenue"`
}

type topProduct struct {
	ProductName string  `json:"product_name"`
	Units       int     `json:"units"`
	Revenue     float64 `json:"revenue"`
}

type marginReport struct {
	PeriodFrom   string          `json:"period_from"`
	PeriodTo     string          `json:"period_to"`
	TotalRevenue float64         `json:"total_revenue"`
	TotalCost    float64         `json:"total_cost"`
	GrossMargin  float64         `json:"gross_margin"`
	MarginPct    float64         `json:"margin_pct"`
	ByProduct    []productMargin `json:"by_product"`
}

type productMargin struct {
	ProductName string  `json:"product_name"`
	Units       int     `json:"units"`
	Revenue     float64 `json:"revenue"`
	Cost        float64 `json:"cost"`
	Margin      float64 `json:"margin"`
	MarginPct   float64 `json:"margin_pct"`
}

// ── P&L types ─────────────────────────────────────────────────────────────────

type pnlReport struct {
	PeriodFrom string `json:"period_from"`
	PeriodTo   string `json:"period_to"`
	// Revenue
	GrossRevenue   float64 `json:"gross_revenue"`
	TotalDiscounts float64 `json:"total_discounts"`
	NetRevenue     float64 `json:"net_revenue"`
	// Cost & Profit
	COGS           float64 `json:"cogs"`
	GrossProfit    float64 `json:"gross_profit"`
	GrossMarginPct float64 `json:"gross_margin_pct"`
	// Sales summary
	TotalSales     int `json:"total_sales"`
	TotalUnitsSold int `json:"total_units_sold"`
	// Investments in period (stock bought)
	TotalInvested float64 `json:"total_invested"`
	UnitsInvested int     `json:"units_invested"`
	// Detailed breakdowns
	ByProduct   []pnlProductRow  `json:"by_product"`
	ByDay       []pnlDayRow      `json:"by_day"`
	ByChannel   []pnlChannelRow  `json:"by_channel"`
	Investments []investmentItem `json:"investments"`
}

type pnlProductRow struct {
	ProductName    string  `json:"product_name"`
	CategoryName   string  `json:"category_name"`
	UnitsSold      int     `json:"units_sold"`
	GrossRevenue   float64 `json:"gross_revenue"`
	TotalDiscounts float64 `json:"total_discounts"`
	NetRevenue     float64 `json:"net_revenue"`
	COGS           float64 `json:"cogs"`
	GrossProfit    float64 `json:"gross_profit"`
	MarginPct      float64 `json:"margin_pct"`
}

type pnlDayRow struct {
	Day         string  `json:"day"`
	NetRevenue  float64 `json:"net_revenue"`
	COGS        float64 `json:"cogs"`
	GrossProfit float64 `json:"gross_profit"`
}

type pnlChannelRow struct {
	Channel    string  `json:"channel"`
	Count      int     `json:"count"`
	NetRevenue float64 `json:"net_revenue"`
}

type investmentItem struct {
	ProductName string  `json:"product_name"`
	SizeName    string  `json:"size_name"`
	ColorName   string  `json:"color_name"`
	Units       int     `json:"units"`
	UnitCost    float64 `json:"unit_cost"`
	TotalCost   float64 `json:"total_cost"`
	Date        string  `json:"date"`
}

// ──────────────────────────────────────────────────────────────────────────────
// Router
// ──────────────────────────────────────────────────────────────────────────────

func ReportHandler(ctx context.Context, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	db := ctx.Value(dto.KeyDatabase).(*sql.DB)
	method := ctx.Value(dto.KeyMethod).(string)
	path := ctx.Value(dto.KeyPath).(string)

	if method == http.MethodOptions {
		return helpers.PreflightResponse(req)
	}

	if method != http.MethodGet {
		return helpers.ErrorResponse(405, "method not allowed", req.Headers["origin"])
	}

	remaining := strings.TrimPrefix(path, "reports")
	remaining = strings.Trim(remaining, "/")

	from, to := parseDateRange(req.QueryStringParameters)

	switch remaining {
	case "sales":
		return buildSalesReport(db, from, to, req)
	case "margin":
		return buildMarginReport(db, from, to, req)
	case "pnl":
		return buildPnLReport(db, from, to, req)
	}
	return helpers.ErrorResponse(404, "report not found", req.Headers["origin"])
}

// parseDateRange reads ?from=YYYY-MM-DD&to=YYYY-MM-DD and defaults to the
// current calendar month when values are absent or unparseable.
func parseDateRange(qp map[string]string) (from, to time.Time) {
	now := time.Now().UTC()

	if f, err := time.Parse("2006-01-02", qp["from"]); err == nil {
		from = time.Date(f.Year(), f.Month(), f.Day(), 0, 0, 0, 0, time.UTC)
	} else {
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	}

	if t, err := time.Parse("2006-01-02", qp["to"]); err == nil {
		to = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.UTC)
	} else {
		to = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)
	}
	return
}

// ──────────────────────────────────────────────────────────────────────────────
// Sales report
// ──────────────────────────────────────────────────────────────────────────────

func buildSalesReport(db *sql.DB, from, to time.Time, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	r := salesReport{
		PeriodFrom:  from.Format("2006-01-02"),
		PeriodTo:    to.Format("2006-01-02"),
		ByChannel:   []channelRow{},
		ByStatus:    []statusRow{},
		ByDay:       []dayRow{},
		TopProducts: []topProduct{},
	}

	// Totals — computed from sale_items (same source as P&L) excluding cancelled/refunded
	var totalDiscount float64
	_ = db.QueryRow(`
		SELECT
		    COUNT(DISTINCT s.id),
		    COALESCE(SUM(si.quantity * si.discount), 0),
		    COALESCE(SUM(si.quantity * (si.unit_price - si.discount)), 0)
		FROM sale_items si
		JOIN sales s ON s.id = si.sale_id
		WHERE s.created_at >= $1 AND s.created_at <= $2
		  AND s.status NOT IN ('cancelled', 'refunded')`,
		from, to,
	).Scan(&r.TotalSales, &totalDiscount, &r.TotalRevenue)
	r.TotalDiscount = totalDiscount
	if r.TotalSales > 0 {
		r.AvgOrder = r.TotalRevenue / float64(r.TotalSales)
	}

	// By channel — from sale_items, excluding cancelled/refunded
	rows, err := db.Query(`
		SELECT COALESCE(s.channel,'sin canal'), COUNT(DISTINCT s.id),
		       COALESCE(SUM(si.quantity * (si.unit_price - si.discount)), 0)
		FROM sale_items si
		JOIN sales s ON s.id = si.sale_id
		WHERE s.created_at >= $1 AND s.created_at <= $2
		  AND s.status NOT IN ('cancelled', 'refunded')
		GROUP BY s.channel
		ORDER BY SUM(si.quantity * (si.unit_price - si.discount)) DESC NULLS LAST`,
		from, to)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var row channelRow
			if rows.Scan(&row.Channel, &row.Count, &row.Revenue) == nil {
				r.ByChannel = append(r.ByChannel, row)
			}
		}
	}

	// By status — counts only from sales header (shows all statuses including cancelled)
	rows2, err := db.Query(`
		SELECT COALESCE(status,'unknown'), COUNT(*), COALESCE(SUM(total),0)
		FROM sales
		WHERE created_at >= $1 AND created_at <= $2
		GROUP BY status
		ORDER BY COUNT(*) DESC`,
		from, to)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var row statusRow
			if rows2.Scan(&row.Status, &row.Count, &row.Revenue) == nil {
				r.ByStatus = append(r.ByStatus, row)
			}
		}
	}

	// By day — from sale_items, excluding cancelled/refunded
	rows3, err := db.Query(`
		SELECT DATE(s.created_at)::text, COUNT(DISTINCT s.id),
		       COALESCE(SUM(si.quantity * (si.unit_price - si.discount)), 0)
		FROM sale_items si
		JOIN sales s ON s.id = si.sale_id
		WHERE s.created_at >= $1 AND s.created_at <= $2
		  AND s.status NOT IN ('cancelled', 'refunded')
		GROUP BY DATE(s.created_at)
		ORDER BY DATE(s.created_at)`,
		from, to)
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var row dayRow
			if rows3.Scan(&row.Day, &row.Count, &row.Revenue) == nil {
				r.ByDay = append(r.ByDay, row)
			}
		}
	}

	// Top products — excluding cancelled/refunded
	rows4, err := db.Query(`
		SELECT p.name,
		       SUM(si.quantity)::int,
		       COALESCE(SUM(si.quantity * (si.unit_price - si.discount)), 0)
		FROM sale_items si
		JOIN sales s             ON s.id  = si.sale_id
		JOIN product_variants pv ON pv.id = si.product_variant_id
		JOIN products p          ON p.id  = pv.product_id
		WHERE s.created_at >= $1 AND s.created_at <= $2
		  AND s.status NOT IN ('cancelled', 'refunded')
		GROUP BY p.id, p.name
		ORDER BY SUM(si.quantity * (si.unit_price - si.discount)) DESC
		LIMIT 10`,
		from, to)
	if err == nil {
		defer rows4.Close()
		for rows4.Next() {
			var row topProduct
			if rows4.Scan(&row.ProductName, &row.Units, &row.Revenue) == nil {
				r.TopProducts = append(r.TopProducts, row)
			}
		}
	}

	return helpers.JsonResponse(200, r, req)
}

// ──────────────────────────────────────────────────────────────────────────────
// P&L report — full profit-and-loss breakdown
// ──────────────────────────────────────────────────────────────────────────────

func buildPnLReport(db *sql.DB, from, to time.Time, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	r := pnlReport{
		PeriodFrom:  from.Format("2006-01-02"),
		PeriodTo:    to.Format("2006-01-02"),
		ByProduct:   []pnlProductRow{},
		ByDay:       []pnlDayRow{},
		ByChannel:   []pnlChannelRow{},
		Investments: []investmentItem{},
	}

	// ── 1. Revenue + COGS totals ─────────────────────────────────────────────
	_ = db.QueryRow(`
		SELECT
		    COUNT(DISTINCT s.id),
		    COALESCE(SUM(si.quantity), 0),
		    COALESCE(SUM(si.quantity * si.unit_price), 0),
		    COALESCE(SUM(si.quantity * si.discount), 0),
		    COALESCE(SUM(
		        si.quantity * COALESCE((
	            SELECT SUM(im.quantity * im.unit_cost) / NULLIF(SUM(im.quantity), 0)
	            FROM inventory_movements im
	            WHERE im.product_variant_id = si.product_variant_id
	              AND im.type = 'IN' AND im.unit_cost IS NOT NULL
	              AND im.created_at <= s.created_at
		        ), 0)
		    ), 0)
		FROM sale_items si
		JOIN sales s ON s.id = si.sale_id
		WHERE s.created_at >= $1 AND s.created_at <= $2
		  AND s.status NOT IN ('cancelled', 'refunded')`,
		from, to,
	).Scan(&r.TotalSales, &r.TotalUnitsSold, &r.GrossRevenue, &r.TotalDiscounts, &r.COGS)

	r.NetRevenue = r.GrossRevenue - r.TotalDiscounts
	r.GrossProfit = r.NetRevenue - r.COGS
	if r.NetRevenue > 0 {
		r.GrossMarginPct = (r.GrossProfit / r.NetRevenue) * 100
	}

	// ── 2. By product ────────────────────────────────────────────────────────
	rows, err := db.Query(`
		SELECT
		    p.name,
		    COALESCE(cat.name, 'Sin categoría'),
		    SUM(si.quantity)::int,
		    COALESCE(SUM(si.quantity * si.unit_price), 0),
		    COALESCE(SUM(si.quantity * si.discount), 0),
		    COALESCE(SUM(
		        si.quantity * COALESCE((
	            SELECT SUM(im.quantity * im.unit_cost) / NULLIF(SUM(im.quantity), 0)
	            FROM inventory_movements im
	            WHERE im.product_variant_id = si.product_variant_id
	              AND im.type = 'IN' AND im.unit_cost IS NOT NULL
	              AND im.created_at <= s.created_at
		        ), 0)
		    ), 0)
		FROM sale_items si
		JOIN sales s              ON s.id  = si.sale_id
		JOIN product_variants pv  ON pv.id = si.product_variant_id
		JOIN products p           ON p.id  = pv.product_id
		LEFT JOIN categories cat  ON cat.id = p.category_id
		WHERE s.created_at >= $1 AND s.created_at <= $2
		  AND s.status NOT IN ('cancelled', 'refunded')
		GROUP BY p.id, p.name, cat.name
		ORDER BY SUM(si.quantity * (si.unit_price - si.discount)) DESC`,
		from, to)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var row pnlProductRow
			if err := rows.Scan(&row.ProductName, &row.CategoryName,
				&row.UnitsSold, &row.GrossRevenue, &row.TotalDiscounts, &row.COGS); err == nil {
				row.NetRevenue = row.GrossRevenue - row.TotalDiscounts
				row.GrossProfit = row.NetRevenue - row.COGS
				if row.NetRevenue > 0 {
					row.MarginPct = (row.GrossProfit / row.NetRevenue) * 100
				}
				r.ByProduct = append(r.ByProduct, row)
			}
		}
	}

	// ── 3. By day ────────────────────────────────────────────────────────────
	rows2, err := db.Query(`
		SELECT
		    DATE(s.created_at)::text,
		    COALESCE(SUM(si.quantity * (si.unit_price - si.discount)), 0),
		    COALESCE(SUM(
		        si.quantity * COALESCE((
	            SELECT SUM(im.quantity * im.unit_cost) / NULLIF(SUM(im.quantity), 0)
	            FROM inventory_movements im
	            WHERE im.product_variant_id = si.product_variant_id
	              AND im.type = 'IN' AND im.unit_cost IS NOT NULL
	              AND im.created_at <= s.created_at
		        ), 0)
		    ), 0)
		FROM sale_items si
		JOIN sales s ON s.id = si.sale_id
		WHERE s.created_at >= $1 AND s.created_at <= $2
		  AND s.status NOT IN ('cancelled', 'refunded')
		GROUP BY DATE(s.created_at)
		ORDER BY DATE(s.created_at)`,
		from, to)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var row pnlDayRow
			if rows2.Scan(&row.Day, &row.NetRevenue, &row.COGS) == nil {
				row.GrossProfit = row.NetRevenue - row.COGS
				r.ByDay = append(r.ByDay, row)
			}
		}
	}

	// ── 4. By channel ────────────────────────────────────────────────────────
	rows3, err := db.Query(`
		SELECT
		    COALESCE(s.channel, 'sin canal'),
		    COUNT(DISTINCT s.id),
		    COALESCE(SUM(si.quantity * (si.unit_price - si.discount)), 0)
		FROM sale_items si
		JOIN sales s ON s.id = si.sale_id
		WHERE s.created_at >= $1 AND s.created_at <= $2
		  AND s.status NOT IN ('cancelled', 'refunded')
		GROUP BY s.channel
		ORDER BY SUM(si.quantity * (si.unit_price - si.discount)) DESC`,
		from, to)
	if err == nil {
		defer rows3.Close()
		for rows3.Next() {
			var row pnlChannelRow
			if rows3.Scan(&row.Channel, &row.Count, &row.NetRevenue) == nil {
				r.ByChannel = append(r.ByChannel, row)
			}
		}
	}

	// ── 5. Investments in period ─────────────────────────────────────────────
	_ = db.QueryRow(`
		SELECT COALESCE(SUM(quantity), 0), COALESCE(SUM(quantity * COALESCE(unit_cost, 0)), 0)
		FROM inventory_movements
		WHERE type = 'IN' AND created_at >= $1 AND created_at <= $2`,
		from, to,
	).Scan(&r.UnitsInvested, &r.TotalInvested)

	rows4, err := db.Query(`
		SELECT
		    p.name, sz.name, c.name,
		    SUM(im.quantity)::int,
		    MAX(im.unit_cost),
		    COALESCE(SUM(im.quantity * COALESCE(im.unit_cost, 0)), 0),
		    MIN(DATE(im.created_at))::text
		FROM inventory_movements im
		JOIN product_variants pv ON pv.id = im.product_variant_id
		JOIN products p          ON p.id  = pv.product_id
		JOIN sizes sz            ON sz.id = pv.size_id
		JOIN colors c            ON c.id  = pv.color_id
		WHERE im.type = 'IN'
		  AND im.created_at >= $1 AND im.created_at <= $2
		GROUP BY p.name, sz.name, c.name
		ORDER BY SUM(im.quantity * COALESCE(im.unit_cost, 0)) DESC`,
		from, to)
	if err == nil {
		defer rows4.Close()
		for rows4.Next() {
			var row investmentItem
			if rows4.Scan(&row.ProductName, &row.SizeName, &row.ColorName,
				&row.Units, &row.UnitCost, &row.TotalCost, &row.Date) == nil {
				r.Investments = append(r.Investments, row)
			}
		}
	}

	return helpers.JsonResponse(200, r, req)
}

// ──────────────────────────────────────────────────────────────────────────────
// Margin report
// ──────────────────────────────────────────────────────────────────────────────

func buildMarginReport(db *sql.DB, from, to time.Time, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	r := marginReport{
		PeriodFrom: from.Format("2006-01-02"),
		PeriodTo:   to.Format("2006-01-02"),
		ByProduct:  []productMargin{},
	}

	rows, err := db.Query(`
		SELECT
		    p.name,
		    SUM(si.quantity)::int                                     AS units,
		    COALESCE(SUM(si.quantity * (si.unit_price - si.discount)), 0) AS revenue,
		    COALESCE(SUM(
		        si.quantity * COALESCE(
	            (SELECT SUM(im.quantity * im.unit_cost) / NULLIF(SUM(im.quantity), 0)
	               FROM inventory_movements im
	              WHERE im.product_variant_id = si.product_variant_id
	                AND im.type = 'IN'
	                AND im.unit_cost IS NOT NULL
	                AND im.created_at <= s.created_at),
		        0)
		    ), 0) AS cost
		FROM sale_items si
		JOIN sales s             ON s.id  = si.sale_id
		JOIN product_variants pv ON pv.id = si.product_variant_id
		JOIN products p          ON p.id  = pv.product_id
		WHERE s.created_at >= $1 AND s.created_at <= $2
		GROUP BY p.id, p.name
		ORDER BY revenue DESC`,
		from, to)
	if err != nil {
		return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
	}
	defer rows.Close()

	for rows.Next() {
		var pm productMargin
		if err := rows.Scan(&pm.ProductName, &pm.Units, &pm.Revenue, &pm.Cost); err != nil {
			return helpers.ErrorResponse(500, err.Error(), req.Headers["origin"])
		}
		pm.Margin = pm.Revenue - pm.Cost
		if pm.Revenue > 0 {
			pm.MarginPct = (pm.Margin / pm.Revenue) * 100
		}
		r.TotalRevenue += pm.Revenue
		r.TotalCost += pm.Cost
		r.ByProduct = append(r.ByProduct, pm)
	}

	r.GrossMargin = r.TotalRevenue - r.TotalCost
	if r.TotalRevenue > 0 {
		r.MarginPct = (r.GrossMargin / r.TotalRevenue) * 100
	}

	return helpers.JsonResponse(200, r, req)
}

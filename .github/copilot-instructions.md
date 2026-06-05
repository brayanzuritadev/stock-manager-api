# Copilot Instructions – StockManager (Go backend)

## Stack
- **Go** (standard library + third-party packages)
- **Gin** HTTP router (used locally/dev); **AWS Lambda + API Gateway** for production
- **PostgreSQL** via `database/sql` (no ORM)
- **Cloudinary** for image upload/delete
- **AWS S3/SDK** via `pkg/awsgo`
- **JWT** for authentication (`internal/middleware/auth.go`)

## Project structure
```
internal/
  handlers/       # Route handlers — one file per domain
  middleware/     # JWT auth middleware
  models/
    domain/       # DB-mapped structs (snake_case JSON tags)
    dto/          # Request/response DTOs
  repositories/   # One file per domain; raw SQL, no ORM
  routes/         # RouteRequest dispatcher (switch on path prefix)
  services/       # Business logic layer between handlers and repos
migrations/       # Numbered SQL migration files
pkg/
  awsgo/          # AWS SDK initialization
  cloudinarypkg/  # Cloudinary helpers
  db/             # postgres.go — DB connection
```

## Conventions

### Handlers (`internal/handlers/`)
- Each handler file receives an `*events.APIGatewayProxyRequest` and returns `(*events.APIGatewayProxyResponse, error)`
- Switch on `req.HTTPMethod` inside each handler
- Use helper `response(statusCode, body)` pattern for JSON responses
- Parse path params from `req.PathParameters`, body from `req.Body`

### Repositories (`internal/repositories/`)
- Accept `*sql.DB` as dependency
- Return domain structs and `error`
- Use `db.QueryRowContext` / `db.QueryContext` with `context.Background()`
- No ORM — write raw SQL
- Named files: `<domain>_repository.go`

### Services (`internal/services/`)
- Thin orchestration layer; call repositories and apply business rules
- Named files: `<domain>_service.go`

### Models
- `domain/` structs map 1-to-1 with DB tables; JSON tags use `snake_case`
- `dto/` structs are request/response shapes; can differ from domain
- Use pointer types (`*string`, `*float64`) for optional/nullable fields

### Routes (`internal/routes/routes.go`)
- Add new route prefix constant + `case` in `RouteRequest` switch
- Path format: `<resource>/<optional-sub-resource>`

## Database
- PostgreSQL; connection string from env var `DATABASE_URL`
- Migrations are plain numbered SQL files in `migrations/`
- New migrations: `00N_description.sql`

### Schema

```sql
CREATE TABLE public.users (
  id            SERIAL PRIMARY KEY,
  name          VARCHAR NOT NULL,
  email         VARCHAR NOT NULL UNIQUE,
  password_hash VARCHAR NOT NULL,
  phone         VARCHAR,
  created_at    TIMESTAMP DEFAULT now(),
  updated_at    TIMESTAMP DEFAULT now(),
  deleted_at    TIMESTAMP          -- soft delete
);

CREATE TABLE public.categories (
  id          SERIAL PRIMARY KEY,
  name        VARCHAR NOT NULL,
  description TEXT,
  created_at  TIMESTAMP DEFAULT now(),
  updated_at  TIMESTAMP DEFAULT now(),
  deleted_at  TIMESTAMP
);

CREATE TABLE public.sizes (
  id         SERIAL PRIMARY KEY,
  name       VARCHAR NOT NULL UNIQUE,
  sort_order INT NOT NULL,
  created_at TIMESTAMP DEFAULT now(),
  deleted_at TIMESTAMP
);

CREATE TABLE public.colors (
  id         SERIAL PRIMARY KEY,
  name       VARCHAR NOT NULL UNIQUE,
  hex_code   VARCHAR,
  created_at TIMESTAMP DEFAULT now(),
  deleted_at TIMESTAMP
);

CREATE TABLE public.products (
  id          SERIAL PRIMARY KEY,
  name        VARCHAR NOT NULL,
  description TEXT,
  sale_price  NUMERIC NOT NULL,
  category_id INT REFERENCES public.categories(id),
  active      BOOLEAN DEFAULT true,
  created_at  TIMESTAMP DEFAULT now(),
  updated_at  TIMESTAMP DEFAULT now(),
  deleted_at  TIMESTAMP
);

CREATE TABLE public.product_variants (
  id             SERIAL PRIMARY KEY,
  product_id     INT NOT NULL REFERENCES public.products(id),
  size_id        INT NOT NULL REFERENCES public.sizes(id),
  color_id       INT NOT NULL REFERENCES public.colors(id),
  stock          INT NOT NULL DEFAULT 0,
  price_override NUMERIC,          -- overrides products.sale_price when set
  created_at     TIMESTAMP DEFAULT now(),
  updated_at     TIMESTAMP DEFAULT now()
  -- UNIQUE (product_id, size_id, color_id) enforced at app level
);

CREATE TABLE public.product_images (
  id         SERIAL PRIMARY KEY,
  product_id INT REFERENCES public.products(id),
  color_id   INT REFERENCES public.colors(id),
  url        TEXT NOT NULL,
  public_id  TEXT NOT NULL,        -- Cloudinary public_id for deletion
  sort_order INT DEFAULT 0,
  created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE public.carts (
  id          SERIAL PRIMARY KEY,
  user_id     INT REFERENCES public.users(id),
  status      VARCHAR DEFAULT 'pending',  -- pending | checkout | completed
  shared_link VARCHAR,
  created_at  TIMESTAMP DEFAULT now(),
  updated_at  TIMESTAMP DEFAULT now(),
  deleted_at  TIMESTAMP
);

CREATE TABLE public.cart_items (
  id                 SERIAL PRIMARY KEY,
  cart_id            INT NOT NULL REFERENCES public.carts(id),
  product_variant_id INT NOT NULL REFERENCES public.product_variants(id),
  quantity           INT NOT NULL,
  unit_price         NUMERIC NOT NULL,
  discount           NUMERIC DEFAULT 0,
  created_at         TIMESTAMP DEFAULT now()
);

CREATE TABLE public.sales (
  id             SERIAL PRIMARY KEY,
  user_id        INT REFERENCES public.users(id),
  total          NUMERIC NOT NULL,
  total_discount NUMERIC DEFAULT 0,
  channel        VARCHAR CHECK (channel IN ('web', 'whatsapp', 'tienda')),
  status         VARCHAR DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'cancelled', 'refunded')),
  created_at     TIMESTAMP DEFAULT now(),
  updated_at     TIMESTAMP DEFAULT now()
);

CREATE TABLE public.sale_items (
  id                 SERIAL PRIMARY KEY,
  sale_id            INT NOT NULL REFERENCES public.sales(id),
  product_variant_id INT NOT NULL REFERENCES public.product_variants(id),
  quantity           INT NOT NULL,
  unit_price         NUMERIC NOT NULL,
  discount           NUMERIC DEFAULT 0,
  created_at         TIMESTAMP DEFAULT now()
);

-- NOTE: table is named inventory_movements (not inventory_entries)
CREATE TABLE public.inventory_movements (
  id                 SERIAL PRIMARY KEY,
  product_variant_id INT NOT NULL REFERENCES public.product_variants(id),
  type               VARCHAR NOT NULL CHECK (type IN ('IN', 'OUT')),
  quantity           INT NOT NULL,
  unit_cost          NUMERIC,       -- purchase cost per unit (IN entries)
  reference_type     VARCHAR,       -- e.g. 'SALE', 'PURCHASE', 'ADJUSTMENT'
  reference_id       INT,           -- FK to sale_id or purchase_id
  created_at         TIMESTAMP DEFAULT now()
);
```

### Key relationships
- `products` → `product_variants` (1-N): every stock unit belongs to a variant
- `product_variants` = product × size × color (unique triplet)
- `product_images` can be scoped to a color via `color_id`
- Sales and inventory movements reference `product_variant_id`, never `product_id` directly
- `inventory_movements.type = 'OUT'` is auto-created when a sale is confirmed
- `sales.channel` values: `'web'`, `'whatsapp'`, `'tienda'`
- `sales.status` values: `'pending'`, `'paid'`, `'cancelled'`, `'refunded'`

## Auth
- Protected endpoints check the JWT via `middleware.ValidateToken()`
- Token contains user ID and role claims

## Error responses
- Always return JSON: `{"error": "message"}`
- HTTP status codes: 200, 201, 400, 401, 403, 404, 409, 500
- Never leak internal errors to the client

## Environment variables
- `DATABASE_URL` – PostgreSQL connection string
- `JWT_SECRET` – signing key
- `CLOUDINARY_URL` – Cloudinary connection
- `AWS_*` – AWS credentials/region

## Do NOT
- Use an ORM
- Store secrets in code
- Return raw `error.Error()` strings directly to the API caller
- Create `.md` summary files after changes

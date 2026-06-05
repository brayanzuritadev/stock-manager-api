package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	//"os"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/joho/godotenv"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/brayanzuritadev/StockManager/internal/middleware"
	"github.com/brayanzuritadev/StockManager/internal/models/dto"
	"github.com/brayanzuritadev/StockManager/internal/repositories"
	"github.com/brayanzuritadev/StockManager/internal/routes"
	"github.com/brayanzuritadev/StockManager/pkg/awsgo"
	"github.com/brayanzuritadev/StockManager/pkg/cloudinarypkg"
	"github.com/brayanzuritadev/StockManager/pkg/db"
	"github.com/gin-gonic/gin"
)

var (
	awsCtx   *awsgo.AWSContext
	database *sql.DB
)

func main() {
	var err error
	godotenv.Load()

	awsCtx, err = awsgo.NewAWSContext()
	if err != nil {
		log.Fatalf("Failed to initialize AWS context: %v", err)
	}

	database, err = db.ConnectDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	lambda.Start(executeLambda)
	//startLocalServer()
}

func executeLambda(ctx context.Context, request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	//path := strings.Replace(request.PathParameters["sfashion"], os.Getenv(awsCtx.UrlPrefix), "", -1)
	path := request.PathParameters["stock-manager"]
	
	ctx = withRequestContext(ctx, request, path, database)
	fmt.Println(path)
	return routes.RouteRequest(path, ctx, &request)
}

func withRequestContext(ctx context.Context, req events.APIGatewayProxyRequest, path string, db *sql.DB) context.Context {
	ctx = context.WithValue(ctx, dto.KeyPath, path)
	ctx = context.WithValue(ctx, dto.KeyMethod, req.HTTPMethod)
	ctx = context.WithValue(ctx, dto.KeyBody, req.Body)
	ctx = context.WithValue(ctx, dto.KeyDatabase, db)
	ctx = context.WithValue(ctx, dto.KeyJWTSign, awsCtx.Secrets.JWTSign)
	//ctx = context.WithValue(ctx, dto.KeyEmail, awsCtx.Secrets.Email)
	//ctx = context.WithValue(ctx, dto.KeyEmailPwd, awsCtx.Secrets.EmailPassword)
	ctx = context.WithValue(ctx, dto.KeyCloudName, awsCtx.Secrets.CloudName)
	ctx = context.WithValue(ctx, dto.KeyCloudAPIKey, awsCtx.Secrets.CloudAPIKey)
	ctx = context.WithValue(ctx, dto.KeyCloudAPISecret, awsCtx.Secrets.CloudAPISecret)
	return ctx
}

func startLocalServer() {
	r := gin.Default()

	// ── CORS middleware ────────────────────────────────────────────────────
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, Accept")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// ── JWT auth middleware ────────────────────────────────────────────────
	r.Use(middleware.JWTAuth(awsCtx.Secrets.JWTSign))

	r.Any("/*path", func(c *gin.Context) {
		// ── Multipart image upload: POST /products/{id}/images/upload ─────
		rawPath := strings.TrimPrefix(c.Param("path"), "/")
		if c.Request.Method == http.MethodPost && isImageUploadPath(rawPath) {
			handleImageUpload(c, rawPath)
			return
		}

		// Copiar headers entrantes al map que espera APIGatewayProxyRequest
		headers := map[string]string{}
		for k, vals := range c.Request.Header {
			if len(vals) > 0 {
				headers[k] = vals[0]
			}
		}
		// Asegurar que la forma canónica "Authorization" esté presente (GetHeader es case-insensitive)
		if ah := c.GetHeader("Authorization"); ah != "" {
			headers["Authorization"] = ah
		}

		// Simular APIGatewayProxyRequest
		// copy query parameters as API Gateway would provide them
		qparams := map[string]string{}
		for k, vals := range c.Request.URL.Query() {
			if len(vals) > 0 {
				qparams[k] = vals[0]
			}
		}
		request := events.APIGatewayProxyRequest{
			Path:       c.Param("path"),
			HTTPMethod: c.Request.Method,
			Body:       getBody(c),
			Headers:    headers,
			PathParameters: map[string]string{
				"lobby": c.Param("path"),
			},
			QueryStringParameters: qparams,
		}

		ctx := withRequestContext(
			context.Background(),
			request,
			rawPath,
			database,
		)

		response, err := routes.RouteRequest(rawPath, ctx, &request)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.Data(response.StatusCode, "application/json", []byte(response.Body))
	})

	r.Run(":3003")
}

func getBody(c *gin.Context) string {
	bodyBytes, _ := io.ReadAll(c.Request.Body)
	return string(bodyBytes)
}

// isImageUploadPath returns true when path matches products/{id}/images/upload.
func isImageUploadPath(path string) bool {
	// expected: products/<id>/images/upload
	segments := strings.Split(strings.Trim(path, "/"), "/")
	return len(segments) == 4 &&
		segments[0] == "products" &&
		segments[2] == "images" &&
		segments[3] == "upload"
}

// handleImageUpload processes multipart/form-data image uploads for a product.
func handleImageUpload(c *gin.Context, path string) {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	productID, err := strconv.Atoi(segments[1])
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "field 'image' is required"})
		return
	}
	defer file.Close()
	_ = header

	cloudClient, err := cloudinarypkg.NewClient(
		awsCtx.Secrets.CloudName,
		awsCtx.Secrets.CloudAPIKey,
		awsCtx.Secrets.CloudAPISecret,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result, err := cloudClient.UploadFile(c.Request.Context(), file, "products")
	if err != nil {
		log.Printf("[cloudinary] upload FAILED for product %d: %v", productID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[cloudinary] upload OK  product=%d public_id=%s url=%s", productID, result.PublicID, result.URL)

	imageRepo := repositories.NewProductImageRepository(database)
	img, err := imageRepo.Add(productID, nil, result.URL, result.PublicID, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, img)
}

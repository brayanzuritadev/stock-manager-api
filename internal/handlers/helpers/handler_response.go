package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// Respuesta JSON con CORS
func JsonResponse(status int, body any, req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	b, err := json.Marshal(body)

	headers := corsHeaders(req.Headers["origin"])

	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf(`{"error": "Failed to serialize response: %v"}`, err),
			Headers:    headers,
		}, nil
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       string(b),
		Headers:    headers,
	}, nil
}

// Respuesta para peticiones OPTIONS
func PreflightResponse(req *events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	headers := corsHeaders(req.Headers["origin"])
	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    headers,
		Body:       "",
	}, nil
}

func ErrorResponse(status int, message string, origin string) (*events.APIGatewayProxyResponse, error) {
	headers := corsHeaders(origin)
	return &events.APIGatewayProxyResponse{
		StatusCode: status,
		Body:       fmt.Sprintf(`{"error": "%s"}`, message),
		Headers:    headers,
	}, nil
}

func corsHeaders(origin string) map[string]string {
	allowedOrigins := []string{
		"http://localhost:4200",
		"https://sfashion-store.netlify.app",
	}

	for _, o := range allowedOrigins {
		if origin == o {
			return map[string]string{
				"Access-Control-Allow-Origin":  origin,
				"Access-Control-Allow-Methods": "GET, POST, PATCH, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers": "Content-Type, Authorization",
				"Content-Type":                 "application/json",
			}
		}
	}

	return map[string]string{
		"Content-Type": "application/json",
	}
}

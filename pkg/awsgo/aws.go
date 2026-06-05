// awsgo/aws_context.go
package awsgo

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

type AWSContext struct {
	Ctx       context.Context
	Cfg       aws.Config
	Secrets   SecretModel
	UrlPrefix string
}

type SecretModel struct {
	JWTSign        string `json:"jwtsign"`
	//Email          string `json:"email"`
	//EmailPassword  string `json:"emailPassword"`
	CloudName      string `json:"cloudName"`
	CloudAPIKey    string `json:"cloudAPIKey"`
	CloudAPISecret string `json:"cloudAPISecret"`
	DatabaseURL     string `json:"databaseURL"`
}

func NewAWSContext() (*AWSContext, error) {
	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx, config.WithDefaultRegion("us-east-1"))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	jwtSign := os.Getenv("JWT_SIGN")
	if jwtSign == "" {
		return nil, fmt.Errorf("JWT_SIGN env var is not set")
	}

	urlPrefix := os.Getenv("UrlPrefix")
	if urlPrefix == "" {
		return nil, fmt.Errorf("UrlPrefix env var is not set")
	}

	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	if cloudName == "" {
		return nil, fmt.Errorf("CLOUDINARY_CLOUD_NAME env var is not set")
	}

	cloudAPIKey := os.Getenv("CLOUDINARY_API_KEY")
	if cloudAPIKey == "" {
		return nil, fmt.Errorf("CLOUDINARY_API_KEY env var is not set")
	}

	cloudAPISecret := os.Getenv("CLOUDINARY_API_SECRET")
	if cloudAPISecret == "" {
		return nil, fmt.Errorf("CLOUDINARY_API_SECRET env var is not set")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL env var is not set")
	}

	return &AWSContext{
		Ctx: ctx,
		Cfg: cfg,
		Secrets: SecretModel{
			JWTSign:        jwtSign,
			//Email:          email,
			//EmailPassword:  emailPassword,
			CloudName:      cloudName,
			CloudAPIKey:    cloudAPIKey,
			CloudAPISecret: cloudAPISecret,
			DatabaseURL:    databaseURL,
		},
		UrlPrefix: urlPrefix,
	}, nil
}

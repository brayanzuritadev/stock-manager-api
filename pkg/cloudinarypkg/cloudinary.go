package cloudinarypkg

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// Client wraps the Cloudinary SDK for easy use across handlers.
type Client struct {
	cld *cloudinary.Cloudinary
}

func NewClient(cloudName, apiKey, apiSecret string) (*Client, error) {
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to init Cloudinary client: %w", err)
	}
	return &Client{cld: cld}, nil
}

// UploadImageResult holds the result of a Cloudinary upload.
type UploadImageResult struct {
	URL      string
	PublicID string
}

// UploadFile uploads any io.Reader to Cloudinary under the given folder.
func (c *Client) UploadFile(ctx context.Context, file interface{}, folder string) (*UploadImageResult, error) {
	resp, err := c.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: folder,
	})
	if err != nil {
		return nil, fmt.Errorf("cloudinary upload failed: %w", err)
	}
	// The SDK v2 does not always surface API errors as Go errors —
	// they arrive in resp.Error.Message instead.
	if resp.Error.Message != "" {
		return nil, fmt.Errorf("cloudinary API error: %s", resp.Error.Message)
	}
	if resp.SecureURL == "" {
		return nil, fmt.Errorf("cloudinary returned empty URL (check cloud name, API key and secret in .env)")
	}
	return &UploadImageResult{
		URL:      resp.SecureURL,
		PublicID: resp.PublicID,
	}, nil
}

// DeleteFile removes a file from Cloudinary by its public ID.
func (c *Client) DeleteFile(ctx context.Context, publicID string) error {
	invalidate := true
	resp, err := c.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID:   publicID,
		Invalidate: &invalidate,
	})
	if err != nil {
		log.Printf("[cloudinary] DeleteFile error publicID=%s: %v", publicID, err)
		return fmt.Errorf("cloudinary delete failed: %w", err)
	}
	// The SDK v2 surfaces API-level errors in resp.Result ("not found", etc.)
	if resp.Result != "ok" {
		log.Printf("[cloudinary] DeleteFile unexpected result publicID=%s result=%s", publicID, resp.Result)
		return fmt.Errorf("cloudinary delete result: %s", resp.Result)
	}
	log.Printf("[cloudinary] DeleteFile OK publicID=%s", publicID)
	return nil
}

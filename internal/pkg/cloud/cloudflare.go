package cloud

import (
	"context"
	"ryzenlo/to2cloud/internal/models"

	"github.com/cloudflare/cloudflare-go"
)

type Cloudflare struct {
	*cloudflare.API
	APIConfig *models.CloudflareAPIConfig
}

var CloudflareInstance *Cloudflare

func GetCloudflare(apiConfig *models.CloudflareAPIConfig) *Cloudflare {
	if CloudflareInstance == nil {
		CloudflareInstance = newCloudflare(apiConfig)
	}
	return CloudflareInstance
}

func newCloudflare(apiConfig *models.CloudflareAPIConfig) *Cloudflare {
	api, err := cloudflare.New(apiConfig.APIKey, apiConfig.Email)
	if err != nil {
		return nil
	}
	return &Cloudflare{api, apiConfig}
}

func (cf *Cloudflare) CheckByCallingAPI(ctx context.Context) error {
	_, err := cf.UserDetails(ctx)
	return err
}

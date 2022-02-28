package cloud

import (
	"context"
	"ryzenlo/to2cloud/internal/models"

	"github.com/vultr/govultr/v2"
	"golang.org/x/oauth2"
)

type Vultr struct {
	*govultr.Client
	APIConfig *models.VultrAPIConfig
}

var VultrInstance *Vultr

func GetVultr(apiConfig *models.VultrAPIConfig) *Vultr {
	if VultrInstance == nil {
		VultrInstance = newVultr(apiConfig)
	}
	return VultrInstance
}

func newVultr(apiConfig *models.VultrAPIConfig) *Vultr {
	config := &oauth2.Config{}
	ctx := context.Background()
	ts := config.TokenSource(ctx, &oauth2.Token{AccessToken: apiConfig.APIKey})
	vultrClient := govultr.NewClient(oauth2.NewClient(ctx, ts))

	return &Vultr{vultrClient, apiConfig}
}

func (v *Vultr) CheckByCallingAPI(ctx context.Context) error {
	_, err := v.Account.Get(ctx)
	return err
}

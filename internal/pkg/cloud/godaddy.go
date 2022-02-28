package cloud

import (
	"context"
	"fmt"
	"ryzenlo/to2cloud/internal/models"
	"net/http"
)

const GODADDY_BASEURI = "https://api.godaddy.com"

type Godaddy struct {
	*http.Client
	APIConfig *models.GodaddyAPIConfig
}

var GodaddyInstance *Godaddy

func GetGodaddyClient(apiConfig *models.GodaddyAPIConfig) *Godaddy {
	if GodaddyInstance == nil {
		GodaddyInstance = newGodaddyClient(apiConfig)
	}
	return GodaddyInstance
}

func newGodaddyClient(apiConfig *models.GodaddyAPIConfig) *Godaddy {
	return &Godaddy{&http.Client{}, apiConfig}
}

func (g *Godaddy) CheckByCallingAPI(ctx context.Context) error {
	url := fmt.Sprintf("%s/v1/domains/available?domain=ryzen.com", GODADDY_BASEURI)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", g.APIConfig.APIKey, g.APIConfig.APISecret))
	resp, err := g.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("something went wrong when calling godaddy api!")
	}
	return nil
}

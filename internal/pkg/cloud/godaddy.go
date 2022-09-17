package cloud

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"ryzenlo/to2cloud/internal/models"
)

const GODADDY_BASEURI = "https://api.godaddy.com"

type GodaddyDomain struct {
	CreatedAt         string   `json:"createdAt,omitempty"`
	DeletedAt         string   `json:"deletedAt,omitempty"`
	Expires           string   `json:"expires,omitempty"`
	Domain            string   `json:"domain,omitempty"`
	DomainId          int64    `json:"domainId,omitempty"`
	Locked            bool     `json:"locked,omitempty"`
	NameServers       []string `json:"nameServers,omitempty"`
	Status            string   `json:"status,omitempty"`
	TransferProtected bool     `json:"transferProtected,omitempty"`
}

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
		return fmt.Errorf("something went wrong when calling godaddy api")
	}
	return nil
}

func (g *Godaddy) ListDomains(ctx context.Context) ([]GodaddyDomain, error) {
	url := fmt.Sprintf("%s/v1/domains?statuses=%s", GODADDY_BASEURI, "ACTIVE")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", g.APIConfig.APIKey, g.APIConfig.APISecret))
	resp, err := g.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("something went wrong when calling godaddy api")
	}
	var domains []GodaddyDomain
	raw, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("something went wrong when reading response from godaddy api")
	}
	err = json.Unmarshal(raw, &domains)
	if err != nil {
		return nil, fmt.Errorf("something went wrong when json decode response from godaddy api!,error msg:%w", err)
	}
	return domains, nil
}

func (g *Godaddy) EditDomain(ctx context.Context, d GodaddyDomain) error {
	url := fmt.Sprintf("%s/v1/domains/%s", GODADDY_BASEURI, d.Domain)
	//
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(&d); err != nil {
		return err
	}
	//
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", fmt.Sprintf("sso-key %s:%s", g.APIConfig.APIKey, g.APIConfig.APISecret))
	//
	req.Header.Add("Content-Type", "application/json")
	resp, err := g.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("something went wrong when calling godaddy api")
	}
	return nil
}

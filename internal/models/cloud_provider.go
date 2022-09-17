package models

import (
	"encoding/json"
	"time"
)

const (
	CLOUDPROVIDER_TYPE_VPS    = "vps"
	CLOUDPROVIDER_TYPE_DOMAIN = "domain"
	CLOUDPROVIDER_TYPE_CDN    = "cdn"
)

const (
	PROVIDER_GODDAY     = "godaddy"
	PROVIDER_VULTR      = "vultr"
	PROVIDER_CLOUDFLARE = "cloudflare"
)

const (
	PROVIDER_API_SUCCESS = 1
	PROVIDER_API_FAILED  = 0
)

const (
	OS_INSTALLED_BY_OSID     = "os_id"
	OS_INSTALLED_BY_SNAPSHOT = "snapshot"
)

var ProviderTypeMap = map[string]bool{
	CLOUDPROVIDER_TYPE_VPS:    true,
	CLOUDPROVIDER_TYPE_DOMAIN: true,
	CLOUDPROVIDER_TYPE_CDN:    true,
}

var ProviderMap = map[string]bool{
	PROVIDER_GODDAY:     true,
	PROVIDER_VULTR:      true,
	PROVIDER_CLOUDFLARE: true,
}

var OSInstallByMap = map[string]bool{
	OS_INSTALLED_BY_OSID:     true,
	OS_INSTALLED_BY_SNAPSHOT: true,
}

type VultrAPIConfig struct {
	APIKey        string `json:"api_key"`
	SSHKeyID      string `json:"ssh_key_id"`
	SSHPrivateKey string `json:"ssh_private_key"`
}

type CloudflareAPIConfig struct {
	APIKey string `json:"api_key"`
	CAKey  string `json:"ca_key"`
	Email  string `json:"email"`
}

type GodaddyAPIConfig struct {
	APIKey    string `json:"api_key"`
	APISecret string `json:"api_secret"`
}

type CloudProvider struct {
	ID            int    `json:"id" gorm:"primaryKey"`
	Account       string `json:"account" gorm:"account" binding:"required"`
	Name          string `json:"name" gorm:"column:name" binding:"required"`
	Type          string `json:"type" gorm:"column:type" binding:"required"`
	APIChecked    int    `json:"api_checked" gorm:"column:api_checked"`
	LastCheckedAt int64  `json:"last_checked_at" gorm:"column:last_checked_at"`
	APIConfig     string `json:"api_config" gorm:"column:api_config"`
	CreateAt      int64  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt     int64  `json:"updated_at" gorm:"column:updated_at"`
}

func (CloudProvider) TableName() string {
	return "cloud_provider"
}

func GetListBy(typeName, name string) []CloudProvider {
	var providers []CloudProvider
	tx := DBClient.Model(&CloudProvider{})
	if typeName != "" {
		tx.Where("type = ?", typeName)
	}
	if name != "" {
		tx.Where("name = ?", name)
	}
	tx.Find(&providers)
	return providers
}

func GetCloudProviderByID(id int) *CloudProvider {
	var cp CloudProvider
	DBClient.First(&cp, id)
	if cp.ID == 0 {
		return nil
	}
	return &cp
}

func GetCloudProviderBy(name, typeName, account string) *CloudProvider {
	var cp CloudProvider
	DBClient.Where("name = ? AND type = ? AND account = ?", name, typeName, account).First(&cp)
	if cp.ID == 0 {
		return nil
	}
	return &cp
}

func AddProvider(cp *CloudProvider) error {
	ts := time.Now().Unix()
	cp.CreateAt = ts
	cp.UpdatedAt = ts
	result := DBClient.Create(cp)
	return result.Error
}

func EditProvider(cp *CloudProvider) error {
	ts := time.Now().Unix()
	cp.UpdatedAt = ts
	result := DBClient.Model(cp).Updates(map[string]interface{}{"account": cp.Account, "api_config": cp.APIConfig, "updated_at": cp.UpdatedAt})
	return result.Error
}

func DelProvider(ID int) error {
	return DBClient.Delete(&CloudProvider{}, ID).Error
}

func UpdateProviderAPICheckStatus(ID, apiChecked int) error {
	ts := time.Now().Unix()
	result := DBClient.Model(&CloudProvider{ID: ID}).Updates(map[string]interface{}{"api_checked": apiChecked, "last_checked_at": ts})
	return result.Error
}

func InProviderTypeList(py string) bool {
	_, yes := ProviderTypeMap[py]
	return yes
}

func InProviderList(p string) bool {
	_, yes := ProviderMap[p]
	return yes
}

func InOSInstallByMapList(py string) bool {
	_, yes := OSInstallByMap[py]
	return yes
}

func NewVultrAPIConfig(raw string) (VultrAPIConfig, error) {
	var apiConfig VultrAPIConfig
	err := json.Unmarshal([]byte(raw), &apiConfig)
	return apiConfig, err
}

func NewCloudflareAPIConfig(raw string) (CloudflareAPIConfig, error) {
	var apiConfig CloudflareAPIConfig
	err := json.Unmarshal([]byte(raw), &apiConfig)
	return apiConfig, err
}

func NewGodaddyAPIConfig(raw string) (GodaddyAPIConfig, error) {
	var apiConfig GodaddyAPIConfig
	err := json.Unmarshal([]byte(raw), &apiConfig)
	return apiConfig, err
}

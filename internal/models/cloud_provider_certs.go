package models

import "time"

type CloudProviderSSLCert struct {
	ID              int    `json:"id" gorm:"primaryKey"`
	CloudProviderID int    `json:"cloud_provider_id" gorm:"column:cloud_provider_id"`
	ZoneID          string `json:"zone_id" gorm:"column:zone_id"`
	LocalRSAKeyID   int    `json:"local_rsa_key_id" gorm:"column:local_rsa_key_id"`
	HostNames       string `json:"host_names" gorm:"column:host_names"`
	ExpiresOn       int64  `json:"expires_on" gorm:"column:expires_on"`
	CertificateID   string `json:"certificate_id" gorm:"column:certificate_id"`
	Certificate     string `json:"certificate" gorm:"column:certificate"`
	CreateAt        int64  `json:"created_at" gorm:"column:created_at"`
}

func (CloudProviderSSLCert) TableName() string {
	return "cloud_provider_ssl_certs"
}

func GetSSLCertList(CloudProviderID int, ZoneID string) []CloudProviderSSLCert {
	var certs []CloudProviderSSLCert
	tx := DBClient.Model(&CloudProviderSSLCert{})
	if CloudProviderID != 0 {
		tx.Where("cloud_provider_id = ?", CloudProviderID)
	}
	if ZoneID != "" {
		tx.Where("zone_id = ?", ZoneID)
	}
	tx.Find(&certs)
	return certs
}

func CreateSSLCert(c *CloudProviderSSLCert) error {
	ts := time.Now().Unix()
	c.CreateAt = ts
	result := DBClient.Create(c)
	return result.Error
}

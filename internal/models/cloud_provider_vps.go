package models

import (
	"encoding/json"
	"fmt"
	"time"
)

type SetupCDNInfo struct {
	CloudProviderID int    `json:"cloud_provider_id"`
	ZoneID          string `json:"zone_id"`
	DNSRecordID     string `json:"dns_record_id"`
	Domain          string `json:"domain"`
}

type CloudProviderVPS struct {
	ID              int    `json:"id" gorm:"primaryKey"`
	CloudProviderID int    `json:"cloud_provider_id" gorm:"column:cloud_provider_id"`
	InstanceID      string `json:"instance_id" gorm:"column:instance_id"`
	LocalRSAKeyID   int    `json:"local_rsa_key_id" gorm:"column:local_rsa_key_id"`
	Status          string `json:"status" gorm:"column:status"`
	CDNInfo         string `json:"cdn_info" gorm:"column:cdn_info"`
	CreateAt        int64  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt       int64  `json:"updated_at" gorm:"column:updated_at"`
}

func (CloudProviderVPS) TableName() string {
	return "cloud_provider_vps"
}

func (v CloudProviderVPS) GetSetupCDNInfo() (SetupCDNInfo, error) {
	var info SetupCDNInfo
	if v.CDNInfo == "" {
		return info, fmt.Errorf("no cdn info setup for the vps")
	}
	if err := json.Unmarshal([]byte(v.CDNInfo), &info); err != nil {
		return info, fmt.Errorf("cannot get cdn info for the vps")
	}
	return info, nil
}

func GetLocalVPSBy(instanceID string) CloudProviderVPS {
	var vps CloudProviderVPS
	DBClient.Where("instance_id = ?", instanceID).First(&vps)
	return vps
}

func EditLocalVPS(cpv *CloudProviderVPS) error {
	ts := time.Now().Unix()
	cpv.UpdatedAt = ts
	toUpdateData := map[string]interface{}{
		"status":     cpv.Status,
		"cdn_info":   cpv.CDNInfo,
		"updated_at": cpv.UpdatedAt,
	}
	result := DBClient.Model(cpv).Updates(toUpdateData)
	return result.Error
}

func CreateLocalVPS(cpv *CloudProviderVPS) error {
	ts := time.Now().Unix()
	cpv.CreateAt = ts
	cpv.UpdatedAt = ts
	result := DBClient.Create(cpv)
	return result.Error
}

func DelVPS(ID int) error {
	return DBClient.Delete(&CloudProviderVPS{}, ID).Error
}

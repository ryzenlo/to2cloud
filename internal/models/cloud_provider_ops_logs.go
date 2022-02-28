package models

import "time"

const (
	APICALL_SUCCESS = 1
	APICALL_FAILED  = 0
)

type CloudProviderOpsLogs struct {
	ID              int    `json:"id" gorm:"primaryKey"`
	CloudProviderID int    `json:"cloud_provider_id" gorm:"column:cloud_provider_id"`
	APIPath         string `json:"api_path" gorm:"column:api_path"`
	APIBody         string `json:"api_body" gorm:"column:api_body"`
	APIResult       int    `json:"api_result" gorm:"column:api_result"`
	APIResponse     string `json:"api_response" gorm:"column:api_response"`
	CreateAt        int64  `json:"created_at" gorm:"column:created_at"`
}

func (CloudProviderOpsLogs) TableName() string {
	return "cloud_provider_ops_logs"
}

func GetOpsLogsBy(cloudProviderID, currPage, pageSize int) ([]CloudProviderOpsLogs, error) {
	var opsLogs []CloudProviderOpsLogs
	var offset = (currPage - 1) * pageSize
	dbResult := DBClient.Where("cloud_provider_id = ?", cloudProviderID).Limit(pageSize).Offset(offset).Order("id desc").Find(&opsLogs)
	if dbResult.Error != nil {
		return nil, dbResult.Error
	}
	return opsLogs, nil
}

func AddProviderOpsLog(log *CloudProviderOpsLogs) error {
	ts := time.Now().Unix()
	log.CreateAt = ts
	result := DBClient.Create(log)
	return result.Error
}

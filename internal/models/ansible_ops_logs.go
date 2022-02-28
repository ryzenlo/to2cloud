package models

import "time"

const (
	ANSIBLE_STATUS_CREATED = "created"
	ANSIBLE_STATUS_RUNNING = "runnng"
	ANSIBLE_STATUS_DONE    = "done"
)

type AnsibleOpsLogs struct {
	ID                int    `json:"id" gorm:"primaryKey"`
	CloudProviderID   int    `json:"cloud_provider_id" gorm:"column:cloud_provider_id"`
	InstanceID        string `json:"instance_id" gorm:"column:instance_id"`
	AnsiblePlaybook   string `json:"ansible_playbook" gorm:"column:ansible_playbook"`
	AnsibleHostConfig string `json:"ansible_host_config" gorm:"column:ansible_host_config"`
	PlayCmd           string `json:"play_cmd" gorm:"column:play_cmd"`
	PlayResult        string `json:"play_result" gorm:"column:play_result"`
	Status            string `json:"status" gorm:"column:status"`
	CreateAt          int64  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt         int64  `json:"updated_at" gorm:"column:updated_at"`
}

func (AnsibleOpsLogs) TableName() string {
	return "ansible_ops_logs"
}

func GetAnsibleOpsLogsByProvider(cloudProviderID int, currPage, pageSize int) []AnsibleOpsLogs {
	var logs []AnsibleOpsLogs
	var offset = (currPage - 1) * pageSize
	DBClient.Where("cloud_provider_id = ?", cloudProviderID).Limit(pageSize).Offset(offset).Order("id desc").Find(&logs)
	return logs
}

func AddAnsibleOpsLog(log *AnsibleOpsLogs) error {
	ts := time.Now().Unix()
	log.CreateAt = ts
	log.UpdatedAt = ts
	result := DBClient.Create(log)
	return result.Error
}

func ReplaceAnsibleOpsLog(log *AnsibleOpsLogs) error {
	ts := time.Now().Unix()
	log.UpdatedAt = ts
	result := DBClient.Save(log)
	return result.Error
}

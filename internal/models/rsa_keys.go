package models

import (
	"encoding/json"
	"time"
)

const KEY_TYPE_SSH = 0
const KEY_TYPE_CSR = 1

type CloudSSHSubject struct {
	CloudProviderID int    `json:"cloud_provider_id"`
	SSHKeyID        string `json:"ssh_key_id"`
}

type RSAKey struct {
	ID              int    `json:"id" gorm:"primaryKey"`
	Name            string `json:"name" gorm:"column:name" binding:"required"`
	Type            int    `json:"type" gorm:"column:type" binding:"required"`
	PrivateKey      string `json:"private_key" gorm:"column:private_key"`
	PublicKey       string `json:"public_key" gorm:"column:public_key"`
	CloudSSHSubject string `json:"cloud_ssh_subject" gorm:"column:cloud_ssh_subject"`
	CsrSubject      string `json:"csr_subject" gorm:"column:csr_subject"`
	CsrCert         string `json:"csr_cert" gorm:"column:csr_cert"`
	CreateAt        int64  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt       int64  `json:"updated_at" gorm:"column:updated_at"`
}

func (RSAKey) TableName() string {
	return "rsa_keys"
}

func (k RSAKey) GetCloudSSHSubject() CloudSSHSubject {
	sub := CloudSSHSubject{}
	json.Unmarshal([]byte(k.CloudSSHSubject), &sub)
	return sub
}

func GetRSAKeyBy(ID int) RSAKey {
	var key RSAKey
	DBClient.First(&key, ID)
	return key
}

func GetRSAKeyListBy(typeName string) []RSAKey {
	var keys []RSAKey
	tx := DBClient.Model(&RSAKey{})
	if typeName != "" {
		tx.Where("type = ?", typeName)
	}
	tx.Find(&keys)
	return keys
}

func CreateRSAKey(k RSAKey) error {
	ts := time.Now().Unix()
	k.CreateAt = ts
	k.UpdatedAt = ts
	result := DBClient.Create(&k)
	return result.Error
}

func UpdateSSHSubject(k RSAKey, sub CloudSSHSubject) error {
	subJsonRaw, err := json.Marshal(sub)
	if err != nil {
		return err
	}
	ts := time.Now().Unix()
	k.UpdatedAt = ts
	toUpdateData := map[string]interface{}{
		"cloud_ssh_subject": string(subJsonRaw),
		"updated_at":        k.UpdatedAt,
	}
	result := DBClient.Model(k).Updates(toUpdateData)
	return result.Error
}

func DeleteRSAKey(ID int) error {
	return DBClient.Delete(&RSAKey{}, ID).Error
}

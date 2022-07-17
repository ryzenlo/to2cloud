package models

import "time"

const KEY_TYPE_SSH = 0
const KEY_TYPE_CSR = 1

type RSAKey struct {
	ID         int    `json:"id" gorm:"primaryKey"`
	Name       string `json:"name" gorm:"column:name" binding:"required"`
	Type       int    `json:"type" gorm:"column:type" binding:"required"`
	PrivateKey string `json:"private_key" gorm:"column:private_key"`
	PublicKey  string `json:"public_key" gorm:"column:public_key"`
	CsrSubject string `json:"csr_subject" gorm:"column:csr_subject"`
	CsrCert    string `json:"csr_cert" gorm:"column:csr_cert"`
	CreateAt   int64  `json:"created_at" gorm:"column:created_at"`
}

func (RSAKey) TableName() string {
	return "rsa_keys"
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
	result := DBClient.Create(&k)
	return result.Error
}

func DeleteRSAKey(ID int) error {
	return DBClient.Delete(&RSAKey{}, ID).Error
}

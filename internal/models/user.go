package models

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"ryzenlo/to2cloud/internal/pkg/log"
)

const (
	USER_STATUS_ACTIVE = iota
	USER_STATUS_INACTIVE
	USER_STATUS_LOCKED
)

const (
	IS_ROOT_NO = iota
	IS_ROOT_YES
)

var ErrHashPwd = errors.New("failed to hash password")
var ErrUsernameExist = errors.New("username is existed")
var ErrAddUser = errors.New("failed to add user")

type User struct {
	ID        int    `gorm:"primaryKey" json:"id"`
	Username  string `gorm:"column:username" json:"username"`
	Password  string `gorm:"column:password" json:"password"`
	Nickname  string `gorm:"column:nickname" json:"nickname"`
	Status    int    `gorm:"column:status" json:"status"`
	IsRoot    int    `gorm:"column:is_root" json:"is_root"`
	CreateAt  int    `gorm:"column:created_at" json:"created_at"`
	UpdatedAt int    `gorm:"column:updated_at" json:"updated_at"`
	LoginAt   int    `gorm:"column:login_at"  json:"login_at"`
}

func (User) TableName() string {
	return "user"
}

func HashPassword(pwd string) string {
	crypted, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return ""
	}
	return string(crypted)
}

func CheckPassword(hashed, pwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(pwd))
	return err == nil
}

func GetUserByName(name string) *User {
	var user User
	DBClient.First(&user, "username = ?", name)
	return &user
}

func GetUserByUserID(userID int) *User {
	var user User
	DBClient.First(&user, userID)
	return &user
}

func AddUser(param *User) error {
	hasedPwd := HashPassword(param.Password)
	if hasedPwd == "" {
		log.Logger.Infoln(ErrHashPwd.Error())
		return ErrHashPwd
	}
	dbUser := GetUserByName(param.Username)
	if dbUser.ID > 0 {
		return ErrUsernameExist
	}
	now := time.Now()
	nowTs := int(now.Unix())
	newUser := &User{
		Username:  param.Username,
		Password:  hasedPwd,
		Nickname:  param.Nickname,
		Status:    USER_STATUS_ACTIVE,
		IsRoot:    IS_ROOT_NO,
		CreateAt:  nowTs,
		UpdatedAt: nowTs,
		LoginAt:   0,
	}
	result := DBClient.Create(&newUser)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

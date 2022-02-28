package http

import (
	"encoding/json"
	"fmt"
	"ryzenlo/to2cloud/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProviderURIParam struct {
	ID int `uri:"id" binding:"required"`
}

type ProviderParam struct {
	Account          string                     `json:"account" gorm:"account" binding:"required"`
	Name             string                     `json:"name" gorm:"column:name" binding:"required"`
	Type             string                     `json:"type" gorm:"column:type" binding:"required"`
	VultrConfig      models.VultrAPIConfig      `json:"vultr_config"`
	CloudflareConfig models.CloudflareAPIConfig `json:"cloudflare_config"`
	Godaddy          models.GodaddyAPIConfig    `json:"godaddy_config"`
}

func GetProviders(c *gin.Context) {
	providerType := c.Query("type")
	providerName := c.Query("name")
	typeList := models.GetListBy(providerType, providerName)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": typeList})
}

func AddProvider(c *gin.Context) {
	var providerParam ProviderParam
	if err := c.ShouldBindJSON(&providerParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	if !models.InProviderTypeList(providerParam.Type) || !models.InProviderList(providerParam.Name) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "no supported cloud"})
		return
	}
	if cp := models.GetCloudProviderBy(providerParam.Name, providerParam.Type, providerParam.Account); cp != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "Such provider exists"})
		return
	}
	cloudProvider := &models.CloudProvider{
		Account: providerParam.Account,
		Name:    providerParam.Name,
		Type:    providerParam.Type,
	}
	apiConfig, err := formCloudConfigFromRequestParam(&providerParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	cloudProvider.APIConfig = apiConfig
	if err := models.AddProvider(cloudProvider); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, SuccessOperationResponse)
}

func EditProvider(c *gin.Context) {
	var uriParam ProviderURIParam
	if err := c.ShouldBindUri(&uriParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	var providerParam ProviderParam
	if err := c.ShouldBindJSON(&providerParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	cloudProvier := models.GetCloudProviderByID(uriParam.ID)
	if cloudProvier == nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "no such cloud provider"})
		return
	}
	//
	apiConfig, err := formCloudConfigFromRequestParam(&providerParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	if cloudProvier.Account != providerParam.Account {
		if cp := models.GetCloudProviderBy(providerParam.Name, providerParam.Type, providerParam.Account); cp != nil {
			c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "Such provider exists"})
			return
		}
		return
	}
	cloudProvier.Account = providerParam.Account
	cloudProvier.APIConfig = apiConfig
	if err := models.EditProvider(cloudProvier); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "Such provider exists"})
		return
	}
	c.JSON(http.StatusOK, SuccessOperationResponse)
}

func DelProvider(c *gin.Context) {
	var uriParam ProviderURIParam
	if err := c.ShouldBindUri(&uriParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	models.DelProvider(uriParam.ID)
	c.JSON(http.StatusOK, SuccessOperationResponse)
}

func formCloudConfigFromRequestParam(param *ProviderParam) (string, error) {
	var config string
	var configBytes []byte
	var err error
	if param.Type == models.CLOUDPROVIDER_TYPE_CDN {
		if param.CloudflareConfig.APIKey == "" || param.CloudflareConfig.Email == "" {
			return config, fmt.Errorf("for cloud flare api, you need to provide apikey and email")
		}
		configBytes, err = json.Marshal(param.CloudflareConfig)
	} else if param.Type == models.CLOUDPROVIDER_TYPE_VPS {
		if param.VultrConfig.APIKey == "" {
			return config, fmt.Errorf("for calling vultr api, you need to provide apikey")
		}
		configBytes, err = json.Marshal(param.VultrConfig)
	} else {
		if param.Godaddy.APIKey == "" || param.Godaddy.APISecret == "" {
			return config, fmt.Errorf("for calling godaddy api, you need to provide apikey and apisecret")
		}
	}
	if err != nil {
		return "", err
	}
	return string(configBytes), nil
}

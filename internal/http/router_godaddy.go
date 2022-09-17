package http

import (
	"context"
	"fmt"
	"net/http"
	"ryzenlo/to2cloud/internal/models"
	"ryzenlo/to2cloud/internal/pkg/cloud"
	"time"

	"github.com/gin-gonic/gin"
)

type GodaddyURIParam struct {
	ID int `uri:"id" binding:"required"`
}

type GodaddyDomainURIParam struct {
	ID     int    `uri:"id" binding:"required"`
	Domain string `uri:"domain" binding:"required"`
}

func checkGodaddyAPI(c *gin.Context) {
	gd, cp, err := getGodaddy(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()
	if err := gd.CheckByCallingAPI(ctx); err != nil {
		models.UpdateProviderAPICheckStatus(cp.ID, models.PROVIDER_API_FAILED)
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": fmt.Sprintf("something went wrong when calling godaddy api: %s", err.Error())})
		return
	}
	models.UpdateProviderAPICheckStatus(cp.ID, models.PROVIDER_API_SUCCESS)
	c.JSON(http.StatusOK, SuccessOperationResponse)
}

func getGodaddyDomains(c *gin.Context) {
	gd, _, err := getGodaddy(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	domains, err := gd.ListDomains(context.Background())
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": domains})
}

func editGodaddyDomain(c *gin.Context) {
	gd, _, err := getGodaddy(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	var reqParam GodaddyDomainURIParam
	if err := c.ShouldBindUri(&reqParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	//
	var domainReq struct {
		NameServers []string `json:"nameServers" binding:"required"`
	}
	//
	if err := c.ShouldBindJSON(&domainReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	gdReq := cloud.GodaddyDomain{
		NameServers: domainReq.NameServers,
	}
	//
	err = gd.EditDomain(context.Background(), gdReq)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, SuccessOperationResponse)
}

func getGodaddy(c *gin.Context) (*cloud.Godaddy, *models.CloudProvider, error) {
	var cloudParam GodaddyURIParam
	if err := c.ShouldBindUri(&cloudParam); err != nil {
		return nil, nil, err
	}
	return getGodaddyProviderAndClient(cloudParam.ID)
}

func getGodaddyProviderAndClient(cloudProviderID int) (*cloud.Godaddy, *models.CloudProvider, error) {
	var gd *cloud.Godaddy
	cloudProvider := models.GetCloudProviderByID(cloudProviderID)
	if cloudProvider == nil || cloudProvider.Name != models.PROVIDER_GODDAY {
		return nil, nil, fmt.Errorf("no such cloud provider")
	}
	var apiConfig models.GodaddyAPIConfig
	var err error
	apiConfig, err = models.NewGodaddyAPIConfig(cloudProvider.APIConfig)
	if err != nil {
		return nil, cloudProvider, fmt.Errorf("something went wrong when creating godaddy client")
	}
	gd = cloud.GetGodaddyClient(&apiConfig)
	if gd == nil {
		return nil, cloudProvider, fmt.Errorf("something went wrong when creating godaddy client")
	}
	return gd, cloudProvider, nil
}

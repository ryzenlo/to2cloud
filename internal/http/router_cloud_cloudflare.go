package http

import (
	"context"
	"fmt"
	"ryzenlo/to2cloud/internal/models"
	"ryzenlo/to2cloud/internal/pkg/cloud"
	"net/http"

	"github.com/cloudflare/cloudflare-go"
	"github.com/gin-gonic/gin"
)

type CloudflareURIParam struct {
	ID int `uri:"id" binding:"required"`
}

type CloudflareZoneURIParam struct {
	ID     int    `uri:"id" binding:"required"`
	ZoneID string `uri:"zone_id" binding:"required"`
}

type UpdateCloudflareDNSURIParam struct {
	CloudflareURIParam
	ZoneID      string `uri:"zone_id" binding:"required"`
	DNSRecordID string `uri:"dns_record_id" binding:"required"`
}

func checkCloudflareAPI(c *gin.Context) {
	cf, cp, err := getCloudflare(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	if err := cf.CheckByCallingAPI(context.Background()); err != nil {
		models.UpdateProviderAPICheckStatus(cp.ID, models.PROVIDER_API_FAILED)
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "something went wrong when calling cloudflare api"})
		return
	}
	//
	models.UpdateProviderAPICheckStatus(cp.ID, models.PROVIDER_API_SUCCESS)
	c.JSON(http.StatusOK, SuccessOperationResponse)
}

func getCloudflareAccounts(c *gin.Context) {
	cf, _, err := getCloudflare(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	user, err := cf.UserDetails(context.Background())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "something went wrong when calling cloudflare api"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": user.Accounts})
}

func getCloudflareZones(c *gin.Context) {
	cf, _, err := getCloudflare(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	zones, err := cf.ListZones(context.Background())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "something went wrong when calling cloudflare api"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": zones})
}

func getCloudflareZoneDNSRecords(c *gin.Context) {
	cf, _, err := getCloudflare(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	var uriParam CloudflareZoneURIParam
	if err := c.ShouldBindUri(&uriParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	rs, err := cf.DNSRecords(context.Background(), uriParam.ZoneID, cloudflare.DNSRecord{Type: "A"})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": rs})
}

func updateCloudflareZoneDNSRecord(c *gin.Context) {
	cf, _, err := getCloudflare(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	var uriParam UpdateCloudflareDNSURIParam
	if err := c.ShouldBindUri(&uriParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	var dnsRecord cloudflare.DNSRecord
	if err := c.ShouldBindJSON(&dnsRecord); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	if err := cf.UpdateDNSRecord(context.Background(), uriParam.ZoneID, uriParam.DNSRecordID, dnsRecord); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": fmt.Sprintf("failed to update dns record, %s", err.Error())})
		return
	}
	c.JSON(http.StatusOK, SuccessOperationResponse)
}

func getCloudflare(c *gin.Context) (*cloud.Cloudflare, *models.CloudProvider, error) {
	var cf *cloud.Cloudflare
	var cloudParam CloudflareURIParam
	if err := c.ShouldBindUri(&cloudParam); err != nil {
		return nil, nil, err
	}
	cloudProvider := models.GetCloudProviderByID(cloudParam.ID)
	if cloudProvider == nil || cloudProvider.Name != models.PROVIDER_CLOUDFLARE {
		return nil, nil, fmt.Errorf("no such cloud provider")
	}
	var apiConfig models.CloudflareAPIConfig
	var err error
	apiConfig, err = models.NewCloudflareAPIConfig(cloudProvider.APIConfig)
	if err != nil {
		return nil, cloudProvider, fmt.Errorf("something went wrong when creating cloudflare client")
	}
	cf = cloud.GetCloudflare(&apiConfig)
	if cf == nil {
		return nil, cloudProvider, fmt.Errorf("something went wrong when creating cloudflare client")
	}
	return cf, cloudProvider, nil
}

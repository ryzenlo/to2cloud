package http

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"ryzenlo/to2cloud/internal/models"
	"ryzenlo/to2cloud/internal/pkg/cloud"

	"encoding/json"

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

type ZoneRequest struct {
	Name      string `json:"name" binding:"required"`
	AccountID string `json:"account_id" binding:"required"`
	Jumpstart bool   `json:"jumpstart"`
	Type      string `json:"type"`
}

type CertRequest struct {
	Hostnames       []string `json:"hostnames"`
	RequestType     string   `json:"request_type"`
	RequestValidity int      `json:"requested_validity"`
	CSR             string   `json:"csr"`
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
	po := cloudflare.PaginationOptions{Page: 1, PerPage: 20}
	accounts, _, err := cf.Accounts(context.Background(), po)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "something went wrong when calling cloudflare api"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": accounts})
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

func createCloudflareZone(c *gin.Context) {
	cf, cloudProvider, err := getCloudflare(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	var zoneReq ZoneRequest
	if err := c.ShouldBindJSON(&zoneReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	if zoneReq.Type == "" {
		zoneReq.Type = "full"
	}
	zoneReq4log, _ := json.Marshal(&zoneReq)
	var zone cloudflare.Zone
	zone, err = cf.CreateZone(context.Background(), zoneReq.Name, false, cloudflare.Account{ID: zoneReq.AccountID}, "full")
	if err != nil {
		logAPICall(c, cloudProvider, models.APICALL_FAILED, string(zoneReq4log), err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": fmt.Sprintf("failed to create zone, %s", err.Error())})
		return
	}
	zoneResponse4log, _ := json.Marshal(&zone)
	logAPICall(c, cloudProvider, models.APICALL_SUCCESS, string(zoneResponse4log), "")
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": zone})
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

func createCloudflareZoneDNSRecord(c *gin.Context) {
	cf, cloudProvider, err := getCloudflare(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	var uriParam CloudflareZoneURIParam
	if err := c.ShouldBindUri(&uriParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	var dnsRecord cloudflare.DNSRecord
	if err := c.ShouldBindJSON(&dnsRecord); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	req4log, _ := json.Marshal(dnsRecord)
	//
	recordResponse, err := cf.CreateDNSRecord(context.Background(), uriParam.ZoneID, dnsRecord)
	if err != nil {
		logAPICall(c, cloudProvider, models.APICALL_FAILED, string(req4log), err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": fmt.Sprintf("failed to update dns record, %s", err.Error())})
		return
	}
	recordResponse4log, _ := json.Marshal(recordResponse)
	logAPICall(c, cloudProvider, models.APICALL_SUCCESS, string(req4log), string(recordResponse4log))
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": recordResponse.Result})
}

func updateCloudflareZoneDNSRecord(c *gin.Context) {
	cf, cloudProvider, err := getCloudflare(c)
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
	req4log, _ := json.Marshal(dnsRecord)
	//
	err = cf.UpdateDNSRecord(context.Background(), uriParam.ZoneID, uriParam.DNSRecordID, dnsRecord)
	if err != nil {
		logAPICall(c, cloudProvider, models.APICALL_FAILED, string(req4log), err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": fmt.Sprintf("failed to update dns record, %s", err.Error())})
		return
	}
	logAPICall(c, cloudProvider, models.APICALL_SUCCESS, string(req4log), "")
	c.JSON(http.StatusOK, SuccessOperationResponse)
}

func getCloudflareCertificates(c *gin.Context) {
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
	opt := cloudflare.OriginCACertificateListOptions{ZoneID: uriParam.ZoneID}
	certs, err := cf.OriginCertificates(context.Background(), opt)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": certs})
}

func createCloudflareCertificate(c *gin.Context) {
	cf, cloudProvider, err := getCloudflare(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	var uriParam CloudflareZoneURIParam
	if err := c.ShouldBindUri(&uriParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	var certReq CertRequest
	if err := c.ShouldBindJSON(&certReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	cert := cloudflare.OriginCACertificate{
		Hostnames:       certReq.Hostnames,
		RequestType:     certReq.RequestType,
		RequestValidity: certReq.RequestValidity,
		CSR:             certReq.CSR,
	}
	req4log, _ := ioutil.ReadAll(c.Request.Body)
	recordResponse, err := cf.CreateOriginCertificate(context.Background(), cert)
	if err != nil {
		logAPICall(c, cloudProvider, models.APICALL_FAILED, string(req4log), err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": fmt.Sprintf("failed to create a certificate, %s", err.Error())})
		return
	}
	logAPICall(c, cloudProvider, models.APICALL_SUCCESS, string(req4log), "")
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": recordResponse})
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

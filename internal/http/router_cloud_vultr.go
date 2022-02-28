package http

import (
	"context"
	"encoding/json"
	"fmt"
	"ryzenlo/to2cloud/configs"
	"ryzenlo/to2cloud/internal/models"
	"ryzenlo/to2cloud/internal/pkg/ansible"
	"ryzenlo/to2cloud/internal/pkg/cloud"
	"ryzenlo/to2cloud/internal/pkg/log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vultr/govultr/v2"
)

type VultrURIParam struct {
	ID int `uri:"id" binding:"required"`
}

type VultrInstanceURIParam struct {
	VultrURIParam
	InstanceID string `uri:"instance_id" binding:"required"`
}

type RunPlayBookJsonParam struct {
	PlaybookName string `json:"playbook_name" binding:"required"`
	ProxyConfig  struct {
		UseProxy string `json:"use_proxy" binding:"required"`
	}
}

type SSHKeyParam struct {
	Name   string `json:"name" binding:"required"`
	SSHKey string `json:"ssh_key" binding:"required"`
}

func checkVultrAPI(c *gin.Context) {
	vl, cp, err := getVultr(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()
	if err := vl.CheckByCallingAPI(ctx); err != nil {
		models.UpdateProviderAPICheckStatus(cp.ID, models.PROVIDER_API_FAILED)
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": fmt.Sprintf("something went wrong when calling vultr api: %s", err.Error())})
		return
	}
	models.UpdateProviderAPICheckStatus(cp.ID, models.PROVIDER_API_SUCCESS)
	c.JSON(http.StatusOK, SuccessOperationResponse)
}

func getVultrSSHKeys(c *gin.Context) {
	vl, _, err := getVultr(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	listOptions := &govultr.ListOptions{PerPage: 20}
	keys, _, err := vl.SSHKey.List(context.Background(), listOptions)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "something went wrong when calling vultr api"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": keys})
}

func addVultrSSHKey(c *gin.Context) {
	vl, cloudProvider, err := getVultr(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	var reqParam SSHKeyParam
	if err := c.ShouldBindJSON(&reqParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	req := &govultr.SSHKeyReq{
		Name:   reqParam.Name,
		SSHKey: reqParam.SSHKey,
	}
	//log call operation
	requestBody, _ := json.Marshal(req)
	sshkey, err := vl.SSHKey.Create(context.Background(), req)
	if err != nil {
		logAPICall(c, cloudProvider, models.APICALL_FAILED, string(requestBody), err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": fmt.Sprintf("something went wrong when calling vultr api, %s", err.Error())})
		return
	}
	apiResponse, _ := json.Marshal(sshkey)
	logAPICall(c, cloudProvider, models.APICALL_SUCCESS, string(requestBody), string(apiResponse))
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": apiResponse})
}

func getVultrSnapshots(c *gin.Context) {
	vl, _, err := getVultr(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	listOptions := &govultr.ListOptions{PerPage: 20}
	snapshots, _, err := vl.Snapshot.List(context.Background(), listOptions)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "something went wrong when calling vultr api"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": snapshots})
}

func createVultrInstanceSnapshot(c *gin.Context) {
	vl, cloudProvider, err := getVultr(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	var vultrParam VultrInstanceURIParam
	if err := c.ShouldBindUri(&vultrParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	req := &govultr.SnapshotReq{
		InstanceID:  vultrParam.InstanceID,
		Description: "",
	}
	snapshot, err := vl.Snapshot.Create(context.Background(), req)
	if err != nil {
		logAPICall(c, cloudProvider, models.APICALL_FAILED, "", err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": fmt.Sprintf("something went wrong when calling vultr api, %s", err.Error())})
		return
	}
	apiResponse, _ := json.Marshal(snapshot)
	logAPICall(c, cloudProvider, models.APICALL_SUCCESS, "", err.Error())
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": apiResponse})
}

func getVultrInstances(c *gin.Context) {
	vl, _, err := getVultr(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	listOptions := &govultr.ListOptions{PerPage: 20}
	instances, _, err := vl.Instance.List(context.Background(), listOptions)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "something went wrong when calling vultr api"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": instances})
}

func createVultrInstance(c *gin.Context) {
	vl, cloudProvider, err := getVultr(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	//
	var instanceReq struct {
		InstalledBy string `json:"installed_by" binding:"required"`
		SnapshotID  string `json:"snapshot_id"`
	}
	if err := c.ShouldBindJSON(&instanceReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	//
	if !models.InOSInstallByMapList(instanceReq.InstalledBy) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "Not supported install method!"})
		return
	}
	instanceOptions := &govultr.InstanceCreateReq{
		OsID:    387, //ubuntu_20.04,defualt installed by os id
		Plan:    "vc2-1c-1gb",
		Region:  "lax",
		Backups: "disabled",
	}
	if instanceReq.InstalledBy == models.OS_INSTALLED_BY_SNAPSHOT {
		// check if snapshot exists
		snapshot, err := vl.Snapshot.Get(context.Background(), instanceReq.SnapshotID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "No such snapshot in vultr"})
			return
		}
		instanceOptions.OsID = 0
		instanceOptions.SnapshotID = snapshot.ID
	}
	//TODO require a lock
	if vl.APIConfig.SSHKeyID != "" {
		instanceOptions.SSHKeys = []string{vl.APIConfig.SSHKeyID}
	}
	//log call operation
	requestBody, _ := json.Marshal(instanceOptions)
	instance, err := vl.Instance.Create(context.Background(), instanceOptions)
	if err != nil {
		logAPICall(c, cloudProvider, models.APICALL_FAILED, string(requestBody), err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": fmt.Sprintf("something went wrong when create vultr client, %s", err.Error())})
		return
	}
	//
	apiResponse, _ := json.Marshal(instance)
	logAPICall(c, cloudProvider, models.APICALL_SUCCESS, string(requestBody), string(apiResponse))
	//
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": instance})
}

func delVultrInstance(c *gin.Context) {
	vl, cloudProvider, err := getVultr(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	var vultrParam VultrInstanceURIParam
	if err := c.ShouldBindUri(&vultrParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	if err := vl.Instance.Delete(context.Background(), vultrParam.InstanceID); err != nil {
		logAPICall(c, cloudProvider, models.APICALL_FAILED, "", err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "something went wrong when calling vultr api"})
		return
	}
	logAPICall(c, cloudProvider, models.APICALL_SUCCESS, "", "")
	c.JSON(http.StatusOK, SuccessOperationResponse)
}

func getRunPlaybookLogs(c *gin.Context) {
	var vultrParam VultrURIParam
	if err := c.ShouldBindUri(&vultrParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	playLogs := models.GetAnsibleOpsLogsByProvider(vultrParam.ID, 1, 20)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": playLogs})
}

func runPlaybookOnVultrInstance(c *gin.Context) {
	vl, cloudProvider, err := getVultr(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	var vultrParam VultrInstanceURIParam
	if err := c.ShouldBindUri(&vultrParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	var req struct {
		PlaybookName string              `json:"playbook_name" binding:"required"`
		ProxyConfig  configs.ProxyConfig `json:"proxy_config"`
	}
	//
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	//
	instance, err := vl.Instance.Get(context.Background(), vultrParam.InstanceID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "something went wrong when calling vultr api"})
		return
	}
	inventory := ansible.Inventory{
		Name:          "vps",
		Host:          instance.MainIP,
		User:          "root",
		SSHPrivateKey: vl.APIConfig.SSHPrivateKey,
	}
	cmd, err := ansible.NewPlayCmd(configs.Conf, req.PlaybookName, inventory, req.ProxyConfig)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": fmt.Sprintf("something went wrong when creating ansible playbook command,%v", err)})
		return
	}
	if err := cmd.CheckPlaybookSyntax(); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": fmt.Sprintf("cannot execute ansible playbook,%v", err)})
		return
	}
	//log run playbook ops
	var playLog = models.AnsibleOpsLogs{
		CloudProviderID:   cloudProvider.ID,
		InstanceID:        instance.ID,
		AnsiblePlaybook:   cmd.GetPlayBookContent(),
		AnsibleHostConfig: cmd.GetInventoryContent(),
		PlayCmd:           cmd.GetFullCmd(),
		Status:            models.ANSIBLE_STATUS_CREATED,
	}
	if err := models.AddAnsibleOpsLog(&playLog); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": fmt.Sprintf("cannot log runing playbook ops,%v", err)})
		return
	}
	//run playbook in another goroutine, and end serving the api, let the client issue its latest ops result
	go func() {
		playLog.Status = models.ANSIBLE_STATUS_RUNNING
		models.ReplaceAnsibleOpsLog(&playLog)
		//run
		runResult, err := cmd.Run()
		playLog.PlayResult = runResult
		playLog.Status = models.ANSIBLE_STATUS_DONE
		if err != nil {
			playLog.PlayResult = err.Error()
			log.Logger.Errorf(fmt.Sprintf("faile to run ansible playbook,%v", err))
		}
		models.ReplaceAnsibleOpsLog(&playLog)
	}()
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"ansible_ops_log_id": playLog.ID}})
}

//log call operation
func logAPICall(c *gin.Context, cp *models.CloudProvider, callStatus int, requestBody, reponse string) {
	callLog := &models.CloudProviderOpsLogs{
		CloudProviderID: cp.ID,
		APIPath:         c.Request.URL.Path,
		APIResult:       callStatus,
		APIBody:         requestBody,
		APIResponse:     reponse,
	}
	models.AddProviderOpsLog(callLog)
}

func getVultrAccount(c *gin.Context) {
	vl, _, err := getVultr(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	account, err := vl.Account.Get(context.Background())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "something went wrong when calling vultr api"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": account})
}

func getVultr(c *gin.Context) (*cloud.Vultr, *models.CloudProvider, error) {
	var vl *cloud.Vultr
	var vultrParam VultrURIParam
	if err := c.ShouldBindUri(&vultrParam); err != nil {
		return nil, nil, fmt.Errorf(fmt.Sprintf("invalid parameter, %s", err.Error()))
	}
	cloudProvider := models.GetCloudProviderByID(vultrParam.ID)
	if cloudProvider == nil || cloudProvider.Name != models.PROVIDER_VULTR {
		return nil, nil, fmt.Errorf("no such cloud provider")
	}
	var apiConfig models.VultrAPIConfig
	var err error
	apiConfig, err = models.NewVultrAPIConfig(cloudProvider.APIConfig)
	if err != nil {
		return nil, cloudProvider, fmt.Errorf("something went wrong when create vultr client")
	}
	//
	vl = cloud.GetVultr(&apiConfig)
	if vl == nil {
		return nil, cloudProvider, fmt.Errorf("something went wrong when create vultr client")
	}
	return vl, cloudProvider, nil
}

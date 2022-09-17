package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"ryzenlo/to2cloud/configs"
	"ryzenlo/to2cloud/internal/models"
	"ryzenlo/to2cloud/internal/pkg/ansible"
	"ryzenlo/to2cloud/internal/pkg/cloud"
	"ryzenlo/to2cloud/internal/pkg/log"
	"ryzenlo/to2cloud/internal/pkg/util"
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
	RSAKeyID int `json:"rsa_key_id" binding:"required"`
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
	RSAKey := models.GetRSAKeyBy(reqParam.RSAKeyID)
	if RSAKey.ID == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "RSA key pair for ssh does not exist."})
		return
	}
	req := &govultr.SSHKeyReq{
		Name:   RSAKey.Name,
		SSHKey: RSAKey.PublicKey,
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
	//
	sub := models.CloudSSHSubject{
		CloudProviderID: cloudProvider.ID,
		SSHKeyID:        sshkey.ID,
	}
	models.UpdateSSHSubject(RSAKey, sub)
	//
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": sshkey})
}

func delVultrSSHKey(c *gin.Context) {
	vl, cloudProvider, err := getVultr(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	var sshKeyURIParam struct {
		ID string `uri:"sshkey_id" binding:"required"`
	}
	if err := c.ShouldBindUri(&sshKeyURIParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	if err := vl.SSHKey.Delete(context.Background(), sshKeyURIParam.ID); err != nil {
		logAPICall(c, cloudProvider, models.APICALL_FAILED, "", err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": fmt.Sprintf("something went wrong when calling vultr api,%s", err.Error())})
		return
	}
	logAPICall(c, cloudProvider, models.APICALL_SUCCESS, "", "")
	c.JSON(http.StatusOK, SuccessOperationResponse)
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
		OSID        int    `json:"os_id" binding:"required"`
		Plan        string `json:"plan" binding:"required"`
		Region      string `json:"region" binding:"required"`
		RSAKeyID    int    `json:"rsa_key_id" binding:"required"`
		InstalledBy string `json:"installed_by" binding:"required"`
		SnapshotID  string `json:"snapshot_id"`
	}
	if err := c.ShouldBindJSON(&instanceReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	//
	RSAKey := models.GetRSAKeyBy(instanceReq.RSAKeyID)
	if RSAKey.ID == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "RSA key pair for ssh does not exist."})
		return
	}
	sshSub := RSAKey.GetCloudSSHSubject()
	if sshSub.CloudProviderID == 0 || sshSub.SSHKeyID == "" {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "SSH key does not exist, please create it in vultr."})
		return
	}
	//
	if !models.InOSInstallByMapList(instanceReq.InstalledBy) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "Not supported install method!"})
		return
	}
	instanceOptions := &govultr.InstanceCreateReq{
		OsID:    instanceReq.OSID,
		Plan:    instanceReq.Plan,
		Region:  instanceReq.Region,
		SSHKeys: []string{sshSub.SSHKeyID},
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
	cpv := &models.CloudProviderVPS{
		CloudProviderID: cloudProvider.ID,
		InstanceID:      instance.ID,
		LocalRSAKeyID:   instanceReq.RSAKeyID,
		Status:          instance.Status,
	}
	models.CreateLocalVPS(cpv)
	//
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": instance})
}

func updateVultrInstance(c *gin.Context) {
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
	var instanceReq struct {
		Label   string              `json:"label" binding:"required"`
		CDNInfo models.SetupCDNInfo `json:"cdn_info" binding:"required"`
	}
	if err := c.ShouldBindJSON(&instanceReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	patchReq := &govultr.InstanceUpdateReq{Label: instanceReq.Label}
	requestBody, _ := json.Marshal(patchReq)
	instance, err := vl.Instance.Update(context.Background(), vultrParam.InstanceID, patchReq)
	if err != nil {
		logAPICall(c, cloudProvider, models.APICALL_FAILED, string(requestBody), err.Error())
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": fmt.Sprintf("something went wrong when calling vultr api,%s", err.Error())})
		return
	}
	apiResponse, _ := json.Marshal(instance)
	logAPICall(c, cloudProvider, models.APICALL_SUCCESS, string(requestBody), string(apiResponse))
	//
	vps := models.GetLocalVPSBy(vultrParam.InstanceID)
	vps.Status = instance.Status
	if vps.ID != 0 {
		rawCDNInfo, err := json.Marshal(instanceReq.CDNInfo)
		if err == nil {
			vps.CDNInfo = string(rawCDNInfo)
		}
	}
	models.EditLocalVPS(&vps)
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
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": fmt.Sprintf("something went wrong when calling vultr api,%s", err.Error())})
		return
	}
	logAPICall(c, cloudProvider, models.APICALL_SUCCESS, "", "")
	c.JSON(http.StatusOK, SuccessOperationResponse)
}

func GetLocalVPS(c *gin.Context) {
	var vultrParam VultrInstanceURIParam
	if err := c.ShouldBindUri(&vultrParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	vps := models.GetLocalVPSBy(vultrParam.InstanceID)
	if vps.ID == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "VPS does not exist!"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "success", "data": vps})
}

func getAnsibleOpsLogs(c *gin.Context) {
	var vultrParam VultrInstanceURIParam
	if err := c.ShouldBindUri(&vultrParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	playLogs := models.GetAnsibleOpsLogsBy(vultrParam.ID, vultrParam.InstanceID, 1, 20)
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
		PlaybookName      string              `json:"playbook_name" binding:"required"`
		PlaybookVariables map[string]string   `json:"playbook_variable"`
		ProxyConfig       configs.ProxyConfig `json:"proxy_config"`
	}
	//
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	//
	localVPS := models.GetLocalVPSBy(vultrParam.InstanceID)
	if localVPS.ID == 0 || localVPS.LocalRSAKeyID == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "RSA key does not exist in the VPS instance."})
		return
	}
	//
	RSAKey := models.GetRSAKeyBy(localVPS.LocalRSAKeyID)
	if RSAKey.ID == 0 {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "RSA key does not exist."})
		return
	}
	//
	instance, err := vl.Instance.Get(context.Background(), vultrParam.InstanceID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "something went wrong when calling vultr api"})
		return
	}
	extraVariable4Ansible := map[string]string{}
	//get domain certs and key files as playbook variables
	var tmpSSLCertFiles *util.TempSSLCertFiles
	cdnInfo, err := localVPS.GetSetupCDNInfo()
	if err == nil {
		extraVariable4Ansible["hostname"] = cdnInfo.Domain
		//
		sslCerts := models.GetSSLCertList(cdnInfo.CloudProviderID, cdnInfo.ZoneID)
		if len(sslCerts) > 0 {
			sslCert := sslCerts[0]
			sslCertRSAKey := models.GetRSAKeyBy(sslCert.LocalRSAKeyID)
			if sslCertRSAKey.ID != 0 {
				tmpSSLCertFiles, err = util.NewTempSSLCertFiles(sslCert.Certificate, sslCertRSAKey.PrivateKey)
				if err == nil {
					extraVariable4Ansible["cert_path"] = tmpSSLCertFiles.GetCertificatePath()
					extraVariable4Ansible["private_key_path"] = tmpSSLCertFiles.GetPrivateKeyPath()
				}
			}
		}
	}

	inventory := ansible.Inventory{
		Name:          "vps",
		Host:          instance.MainIP,
		User:          "root",
		SSHPrivateKey: RSAKey.PrivateKey,
	}
	cmd, err := ansible.NewPlayCmd(configs.Conf, req.PlaybookName, inventory, req.ProxyConfig, extraVariable4Ansible)
	if err != nil {
		//
		util.RemoveAllSSLTmpFiles(tmpSSLCertFiles)
		//
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": fmt.Sprintf("something went wrong when creating ansible playbook command,%v", err)})
		return
	}
	if err := cmd.CheckPlaybookSyntax(); err != nil {
		//
		util.RemoveAllSSLTmpFiles(tmpSSLCertFiles)
		//
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": fmt.Sprintf("cannot execute ansible playbook,%v", err)})
		return
	}
	//log run playbook ops
	var playLog = models.AnsibleOpsLogs{
		CloudProviderID:        cloudProvider.ID,
		InstanceID:             instance.ID,
		AnsiblePlaybookName:    req.PlaybookName,
		AnsiblePlaybookContent: cmd.GetPlayBookContent(),
		AnsibleExtraVariables:  cmd.GetAnsibleExtraVariables(),
		AnsibleHostConfig:      cmd.GetInventoryContent(),
		PlayCmd:                cmd.GetFullCmd(),
		Status:                 models.ANSIBLE_STATUS_CREATED,
	}
	if err := models.AddAnsibleOpsLog(&playLog); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": fmt.Sprintf("cannot log runing playbook ops,%v", err)})
		return
	}
	//run playbook in another goroutine, and end serving the api, let the client issue its latest ops result
	go func() {
		defer cmd.Clean()
		defer util.RemoveAllSSLTmpFiles(tmpSSLCertFiles)
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

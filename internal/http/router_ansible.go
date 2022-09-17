package http

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"ryzenlo/to2cloud/configs"
	"ryzenlo/to2cloud/internal/pkg/util"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-ping/ping"
)

type AnsiblePlayBook struct {
	Filename string `json:"filename"`
}

type AnsiblePlayBooks []AnsiblePlayBook

func getAnsiblePlayBooks(c *gin.Context) {
	playbookDirPath := configs.Conf.Ansible.DirPath
	if _, err := os.Stat(playbookDirPath); os.IsNotExist(err) {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "not ansible playbooks on the server."})
		return
	}
	files, err := ioutil.ReadDir(playbookDirPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": fmt.Sprintf("failed to get ansible playbooks,%w.", err)})
		return
	}
	var playbooks AnsiblePlayBooks
	for _, v := range files {
		if v.IsDir() {
			continue
		}
		playbooks = append(playbooks, AnsiblePlayBook{
			Filename: v.Name(),
		})
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": playbooks})
}

func getAnsiblePlayBook(c *gin.Context) {
	var uriParam struct {
		FileName string `uri:"filename" binding:"required"`
	}
	if err := c.ShouldBindUri(&uriParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	fullPath := fmt.Sprintf("%s/%s", configs.Conf.Ansible.DirPath, uriParam.FileName)
	if _, err := os.Stat(fullPath); uriParam.FileName == "" || os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"code": 1, "msg": "Not Found"})
		return
	}
	fileContent, _ := ioutil.ReadFile(fullPath)
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "Success", "data": string(fileContent)})
}

func pingIP(c *gin.Context) {
	var ipParam struct {
		IP string `uri:"ip" binding:"required"`
	}
	if err := c.ShouldBindUri(&ipParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	pinger, err := ping.NewPinger(ipParam.IP)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	pinger.SetPrivileged(true)
	pinger.Count = 2
	pinger.Timeout = time.Second * 3
	err = pinger.Run()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	report := pinger.Statistics()
	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "", "data": report})
}

func checkSSHConection(c *gin.Context) {
	//
	var req struct {
		IP            string `json:"ip" binding:"required"`
		Port          string `json:"port"`
		Username      string `json:"username"`
		RSAPrivateKey string `json:"rsa_private_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	keyPair := util.RSAKeyPair{PrivateKey: req.RSAPrivateKey}
	err := util.CheckSSHConection(req.IP, req.Port, req.Username, time.Second*3, keyPair)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, SuccessOperationResponse)
}

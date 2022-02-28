package http

import (
	"github.com/gin-gonic/gin"
)

var SuccessOperationResponse = gin.H{"code": 0, "msg": "success"}

func setupRoutes(r *gin.Engine) {
	r.POST("/login", userLogin)
	//
	needAuth := r.Group("/")
	needAuth.Use(isLogin())
	{
		//
		needRoot := needAuth.Group("")
		needRoot.Use(isRootUser())
		{
			needRoot.POST("/user", AddUser)
		}
		//
		needAuth.GET("/cloud_providers", GetProviders)
		needAuth.POST("/cloud_provider", AddProvider)
		needAuth.PUT("/cloud_provider/:id", EditProvider)
		needAuth.DELETE("/cloud_provider/:id", DelProvider)
		//
		//vultr
		needAuth.GET("/cloud_provider/:id/vultr/check", checkVultrAPI)
		needAuth.GET("/cloud_provider/:id/vultr/account", getVultrAccount)
		needAuth.GET("/cloud_provider/:id/vultr/instances", getVultrInstances)
		needAuth.GET("/cloud_provider/:id/vultr/sshkeys", getVultrSSHKeys)
		needAuth.POST("/cloud_provider/:id/vultr/sshkey", addVultrSSHKey)
		needAuth.GET("/cloud_provider/:id/vultr/snapshots", getVultrSnapshots)
		needAuth.POST("/cloud_provider/:id/vultr/instance", createVultrInstance)
		needAuth.POST("/cloud_provider/:id/vultr/instance/:instance_id/snapshot", createVultrInstanceSnapshot)
		needAuth.DELETE("/cloud_provider/:id/vultr/instance/:instance_id", delVultrInstance)
		//run ansible
		needAuth.GET("/cloud_provider/:id/vultr/ansible/ops_logs", getRunPlaybookLogs)
		needAuth.POST("/cloud_provider/:id/vultr/instance/:instance_id/ansible/ops", runPlaybookOnVultrInstance)
		//cloudflare
		needAuth.GET("/cloud_provider/:id/cloudflare/check", checkCloudflareAPI)
		needAuth.GET("/cloud_provider/:id/cloudflare/accounts", getCloudflareAccounts)
		needAuth.GET("/cloud_provider/:id/cloudflare/zones", getCloudflareZones)
		needAuth.GET("/cloud_provider/:id/cloudflare/zones/:zone_id/dns_records", getCloudflareZoneDNSRecords)
		needAuth.PATCH("/cloud_provider/:id/cloudflare/zones/:zone_id/dns_records/:dns_record_id", updateCloudflareZoneDNSRecord)
	}
}

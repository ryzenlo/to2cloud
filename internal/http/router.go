package http

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var SuccessOperationResponse = gin.H{"code": 0, "msg": "success"}

func setupRoutes(r *gin.Engine) {
	r.POST("/login", userLogin)
	//
	r.Use(cors.Default())
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
		needAuth.GET("/user", GetUser)
		//
		needAuth.GET("/cloud_providers", GetProviders)
		needAuth.POST("/cloud_provider", AddProvider)
		needAuth.PUT("/cloud_provider/:id", EditProvider)
		needAuth.DELETE("/cloud_provider/:id", DelProvider)
		//
		needAuth.GET("/rsa_keys", GetRSAKeys)
		needAuth.POST("/rsa_key", CreateRSAKey)
		needAuth.DELETE("/rsa_key/:id", DeleteRSAKey)
		//
		//vultr
		needAuth.GET("/cloud_provider/:id/vultr/check", checkVultrAPI)
		needAuth.GET("/cloud_provider/:id/vultr/account", getVultrAccount)
		needAuth.GET("/cloud_provider/:id/vultr/instances", getVultrInstances)
		needAuth.POST("/cloud_provider/:id/vultr/instance", createVultrInstance)
		needAuth.PATCH("/cloud_provider/:id/vultr/instance/:instance_id", updateVultrInstance)
		needAuth.GET("/cloud_provider/:id/vultr/sshkeys", getVultrSSHKeys)
		needAuth.POST("/cloud_provider/:id/vultr/sshkey", addVultrSSHKey)
		needAuth.DELETE("/cloud_provider/:id/vultr/sshkey/:sshkey_id", delVultrSSHKey)
		needAuth.GET("/cloud_provider/:id/vultr/snapshots", getVultrSnapshots)
		needAuth.POST("/cloud_provider/:id/vultr/instance/:instance_id/snapshot", createVultrInstanceSnapshot)
		needAuth.DELETE("/cloud_provider/:id/vultr/instance/:instance_id", delVultrInstance)
		//run ansible
		needAuth.GET("/cloud_provider/:id/vultr/ansible/ops_logs", getRunPlaybookLogs)
		needAuth.POST("/cloud_provider/:id/vultr/instance/:instance_id/ansible/ops", runPlaybookOnVultrInstance)
		//cloudflare
		needAuth.GET("/cloud_provider/:id/cloudflare/check", checkCloudflareAPI)
		needAuth.GET("/cloud_provider/:id/cloudflare/accounts", getCloudflareAccounts)
		needAuth.GET("/cloud_provider/:id/cloudflare/zones", getCloudflareZones)
		needAuth.POST("/cloud_provider/:id/cloudflare/zone", createCloudflareZone)
		needAuth.GET("/cloud_provider/:id/cloudflare/zones/:zone_id/dns_records", getCloudflareZoneDNSRecords)
		needAuth.POST("/cloud_provider/:id/cloudflare/zones/:zone_id/dns_records", createCloudflareZoneDNSRecord)
		needAuth.PATCH("/cloud_provider/:id/cloudflare/zones/:zone_id/dns_records/:dns_record_id", updateCloudflareZoneDNSRecord)
		//
		needAuth.GET("/cloud_provider/:id/cloudflare/zones/:zone_id/certificates", getCloudflareCertificates)
		needAuth.POST("/cloud_provider/:id/cloudflare/zones/:zone_id/certificates", createCloudflareCertificate)
	}
}

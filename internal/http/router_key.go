package http

import (
	"encoding/json"
	"net/http"
	"ryzenlo/to2cloud/internal/models"
	"ryzenlo/to2cloud/internal/pkg/util"

	"github.com/gin-gonic/gin"
)

type CreateKeyParam struct {
	Name         string            `json:"name"  binding:"required"`
	Type         int               `json:"type"`
	KeyBits      int               `json:"key_bits"`
	SubjectParam map[string]string `json:"csr_subject_param"`
}

type KeyURIParam struct {
	ID int `uri:"id" binding:"required"`
}

func GetRSAKeys(c *gin.Context) {
	typeName := c.Query("type")
	list := models.GetRSAKeyListBy(typeName)
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": list})
}

func CreateRSAKey(c *gin.Context) {
	var param CreateKeyParam
	if err := c.ShouldBindJSON(&param); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	k := models.RSAKey{
		Name: param.Name,
		Type: param.Type,
	}
	CSR := false
	if param.Type == models.KEY_TYPE_CSR {
		CSR = true
	}
	//
	param.KeyBits = 2048
	keyPair, err := util.GeneKeyPair(param.KeyBits, CSR, param.SubjectParam)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": err.Error()})
		return
	}

	k.PrivateKey = keyPair.PrivateKey
	k.PublicKey = keyPair.PublicKey

	if param.Type == models.KEY_TYPE_CSR {
		raw, _ := json.Marshal(&param.SubjectParam)
		k.CsrSubject = string(raw)
		k.CsrCert = keyPair.CSRCert
	}

	if err := models.CreateRSAKey(k); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, SuccessOperationResponse)
}

func DeleteRSAKey(c *gin.Context) {
	var URIParam KeyURIParam
	if err := c.ShouldBindUri(&URIParam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
	}
	if err := models.DeleteRSAKey(URIParam.ID); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 1, "msg": err.Error()})
	}
	c.JSON(http.StatusOK, SuccessOperationResponse)
}

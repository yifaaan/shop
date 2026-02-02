package handler

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"shop/oss_web/global"

	"github.com/gin-gonic/gin"
)

// Sign 生成阿里云OSS直传签名
func Sign(ctx *gin.Context) {
	ossCfg := global.ServerConfig.OSSConfig
	accessKeyID := ossCfg.AccessKeyID
	accessKeySecret := ossCfg.AccessKeySecret
	if accessKeyID == "" || accessKeySecret == "" {
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "OSS配置缺失"})
		return
	}

	bucket := ossCfg.Bucket
	if bucket == "" {
		bucket = "lyf-shop-files"
	}
	endpoint := strings.TrimSpace(ossCfg.Endpoint)
	if endpoint == "" {
		endpoint = "oss-cn-hangzhou.aliyuncs.com"
	}
	host := "https://" + bucket + "." + endpoint

	dir := strings.TrimSpace(ossCfg.UploadDir)
	if dir == "" {
		dir = "uploads/"
	}
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}
	dir = dir + time.Now().Format("2006/01/02/") // 按日期分目录

	expireSeconds := int64(30 * 60) // 30分钟
	expireTime := time.Now().UTC().Add(time.Duration(expireSeconds) * time.Second)
	expireISO := expireTime.Format("2006-01-02T15:04:05.000Z")

	policy := map[string]any{
		"expiration": expireISO,
		"conditions": []any{
			map[string]string{"bucket": bucket},
			[]any{"starts-with", "$key", dir},
			[]any{"content-length-range", 0, 1024 * 1024 * 1024},
		},
	}
	policyBytes, _ := json.Marshal(policy)
	policyBase64 := base64.StdEncoding.EncodeToString(policyBytes)
	h := hmac.New(sha1.New, []byte(accessKeySecret))
	_, _ = h.Write([]byte(policyBase64))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	ctx.JSON(http.StatusOK, gin.H{
		"accessid":  accessKeyID,
		"policy":    policyBase64,
		"signature": signature,
		"dir":       dir,
		"host":      host,
		"expire":    expireTime.Unix(),
		"bucket":    bucket,
	})
}

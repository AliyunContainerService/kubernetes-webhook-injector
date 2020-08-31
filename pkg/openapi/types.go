package openapi

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
)

type AKInfo struct {
	AccessKeyId     string `json:"access.key.id"`
	AccessKeySecret string `json:"access.key.secret"`
	SecurityToken   string `json:"security.token"`
	Expiration      string `json:"expiration"`
	Keyring         string `json:"keyring"`
}

//type AK struct {
//	AccessKeyId     string
//	AccessKeySecret string
//}
//
//type STS struct {
//}

type SecurityGroupOperator struct {
	*ecs.Client
}

type RdsWhitelistOperator struct {
	*rds.Client
}

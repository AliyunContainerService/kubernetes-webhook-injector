package openapi

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/r-kvstore"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
)

type AKInfo struct {
	AccessKeyId     string `json:"access.key.id"`
	AccessKeySecret string `json:"access.key.secret"`
	SecurityToken   string `json:"security.token"`
	Expiration      string `json:"expiration"`
	Keyring         string `json:"keyring"`
}

type SecurityGroupOperator struct {
	*ecs.Client
}

type RdsWhitelistOperator struct {
	*rds.Client
}

type RedisWhitelistOperator struct {
	*r_kvstore.Client
}

type SLBAccessControlPolicyOperator struct {
	*slb.Client
}

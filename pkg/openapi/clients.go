package openapi

import (
	"fmt"
	"log"
	"os"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/endpoints"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	redis "github.com/aliyun/alibaba-cloud-sdk-go/services/r-kvstore"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
)

var (
	RegionID string
)

func init() {
	r, ok := os.LookupEnv("REGION_ID")
	if !ok {
		log.Println("env REGION_ID is not set")
	}
	RegionID = r

	endpoints.AddEndpointMapping(RegionID, "Ecs", fmt.Sprintf("ecs-vpc.%s.aliyuncs.com", RegionID))
	endpoints.AddEndpointMapping(RegionID, "Rds", fmt.Sprintf("rds-vpc.%s.aliyuncs.com", RegionID))
}

func getSDKClient(authInfo *AKInfo) (*sdk.Client, error) {

	if authInfo.SecurityToken != "" {
		return sdk.NewClientWithStsToken(RegionID, authInfo.AccessKeyId, authInfo.AccessKeySecret, authInfo.SecurityToken)
	} else {
		return sdk.NewClientWithAccessKey(RegionID, authInfo.AccessKeyId, authInfo.AccessKeySecret)
	}
}

func GetSecurityGroupOperator(authInfo *AKInfo) (*SecurityGroupOperator, error) {
	sdkCli, err := getSDKClient(authInfo)
	if err != nil {
		return nil, err
	}

	return &SecurityGroupOperator{Client: &ecs.Client{Client: *sdkCli}}, nil
}

func GetRdsWhitelistOperator(authInfo *AKInfo) (*RdsWhitelistOperator, error) {
	sdkCli, err := getSDKClient(authInfo)
	if err != nil {
		return nil, err
	}

	return &RdsWhitelistOperator{Client: &rds.Client{Client: *sdkCli}}, nil
}

func GetRedisWhiteListOperator(authInfo *AKInfo) (*RedisWhitelistOperator, error) {
	sdkCli, err := getSDKClient(authInfo)
	if err != nil {
		return nil, err
	}

	return &RedisWhitelistOperator{
		Client: &redis.Client{
			Client: *sdkCli,
		},
	}, nil
}

func GetSLBAccessControlPolicyOperator(authInfo *AKInfo) (*SLBAccessControlPolicyOperator, error) {
	sdkCli, err := getSDKClient(authInfo)
	if err != nil {
		return nil, err
	}

	return &SLBAccessControlPolicyOperator{
		Client: &slb.Client{
			Client: *sdkCli,
		},
	}, nil
}

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
	PrivateRegion := []string{"cn-qingdao", "cn-beijing", "cn-hangzhou", "cn-shanghai", "cn-shenzhen", "cn-hongkong", "ap-southeast-1", "us-east-1", "us-west-1", "cn-shanghai-finance-1"}
	SpecialRegion := []string{"cn-heyuan", "cn-hangzhou-finance", "cn-shenzhen-finance-1"}

	r, ok := os.LookupEnv("REGION_ID")
	if !ok {
		log.Println("env REGION_ID is not set")
	}
	RegionID = r

	endpoints.AddEndpointMapping(RegionID, "Rds", fmt.Sprintf("rds.%s.aliyuncs.com", RegionID))

	for _, pr := range PrivateRegion {
		if RegionID == pr {
			endpoints.AddEndpointMapping(RegionID, "Rds", fmt.Sprintf("rds-vpc.%s.aliyuncs.com", RegionID))
		}
	}

	for _, sr := range SpecialRegion {
		if RegionID == sr {
			endpoints.AddEndpointMapping(RegionID, "Rds", fmt.Sprintf("rds.aliyuncs.com"))
		}
	}

	endpoints.AddEndpointMapping(RegionID, "Ecs", fmt.Sprintf("ecs-vpc.%s.aliyuncs.com", RegionID))
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

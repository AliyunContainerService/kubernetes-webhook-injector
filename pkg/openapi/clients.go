package openapi

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"log"
	"os"
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

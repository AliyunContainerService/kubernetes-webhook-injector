package main

import (
	"github.com/AliyunContainerService/kubernetes-webhook-injector/pkg/openapi"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins/utils"
	_ "github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/jessevdk/go-flags"
	"log"
	"strings"
)

type Options struct {
	SecurityGroupIDs string `long:"security_group_id" required:"true"`
	RegionId         string `long:"region_id" required:"true"`
	RuleId           string `long:"permission_id" required:"true"`
	AccessKeyID      string `long:"access_key_id" required:"true"`
	AccessKeySecret  string `long:"access_key_secret" required:"true"`
	StsToken         string `long:"sts_token"`
	ToDelete         bool   `long:"delete"`
}

var (
	opt      Options
	sgClient *openapi.SecurityGroupOperator
	sgIDs    []string
)

func init() {
	_, err := flags.Parse(&opt)
	if err != nil {
		log.Fatal(err)
	}
	sgIDs = strings.Split(opt.SecurityGroupIDs, ",")

	//TODO: 在这里检查STS是否设置，如果有，从STS获取 security_group 客户端
	//opt.AccessKeyID, opt.AccessKeySecret = utils.MustGetAk()

	authInfo := &openapi.AKInfo{
		AccessKeyId:     opt.AccessKeyID,
		AccessKeySecret: opt.AccessKeySecret,
		SecurityToken:   opt.StsToken,
	}

	sgClient, err = openapi.GetSecurityGroupOperator(authInfo)
	if err != nil {
		log.Fatal(err)
	}

}

func main() {

	for _, sgID := range sgIDs {

		//创建rule
		ipAddr, err := utils.ExternalIP()
		if err != nil {
			log.Fatalf("Cannot find pod's external IP: %v", err)
		}
		log.Printf("Creating permission %s of sg %s in region %s\n", opt.RuleId, sgID, opt.RegionId)
		if err := sgClient.CreatePermission(sgID, ipAddr, opt.RuleId); err != nil {
			log.Fatalf("Failed to create permission: %v", err)
		}
		log.Println("Permission created")
	}

}

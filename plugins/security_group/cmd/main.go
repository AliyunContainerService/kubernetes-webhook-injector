package main

import (
	"fmt"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/pkg/openapi"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins/utils"
	_ "github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/jessevdk/go-flags"
	"log"
	"os"
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

const terminationLog = "/dev/termination-log"

var (
	opt Options
)

func main() {
	msgLog, err := os.Create(terminationLog)
	if err != nil {
		log.Fatal(err)
	}
	defer msgLog.Close()
	defer msgLog.Sync()

	_, err = flags.Parse(&opt)
	if err != nil {
		msgLog.WriteString(err.Error())
		log.Fatal(err)
	}
	sgIDs := strings.Split(opt.SecurityGroupIDs, ",")

	authInfo := &openapi.AKInfo{
		AccessKeyId:     opt.AccessKeyID,
		AccessKeySecret: opt.AccessKeySecret,
		SecurityToken:   opt.StsToken,
	}

	sgClient, err := openapi.GetSecurityGroupOperator(authInfo)
	if err != nil {
		msgLog.WriteString(err.Error())
		log.Fatal(err)
	}

	for _, sgID := range sgIDs {

		//创建rule
		ipAddr, err := utils.ExternalIP()
		if err != nil {
			msg := fmt.Sprintf("Cannot find pod's external IP: %v", err)
			msgLog.WriteString(msg)
			log.Fatalf(msg)
		}
		if err := sgClient.CreatePermission(sgID, ipAddr, opt.RuleId); err != nil {
			msg := fmt.Sprintf("Failed to create permission: %v", err)
			msgLog.WriteString(msg)
			log.Fatalf(msg)
		}
		msg := fmt.Sprintf("Created permission %s of sg %s in region %s\n", opt.RuleId, sgID, opt.RegionId)
		log.Printf(msg)
	}
}

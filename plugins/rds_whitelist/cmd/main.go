package main

import (
	"fmt"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/pkg/openapi"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins/utils"
	"github.com/jessevdk/go-flags"
	"log"
	"os"
	"strings"
)

type Options struct {
	RdsIDs          string `long:"rds_id" required:"true"`
	RegionId        string `long:"region_id" required:"true"`
	WhiteListName   string `long:"white_list_name" required:"true"`
	AccessKeyID     string `long:"access_key_id" required:"true"`
	AccessKeySecret string `long:"access_key_secret" required:"true"`
	StsToken        string `long:"sts_token"`
	ToDelete        bool   `long:"delete"`
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
	rdsIDs := strings.Split(opt.RdsIDs, ",")
	opt.WhiteListName = openapi.RefactorRdsWhitelistName(opt.WhiteListName)

	authInfo := &openapi.AKInfo{
		AccessKeyId:     opt.AccessKeyID,
		AccessKeySecret: opt.AccessKeySecret,
		SecurityToken:   opt.StsToken,
	}

	rdsClient, err := openapi.GetRdsWhitelistOperator(authInfo)
	if err != nil {
		msgLog.WriteString(err.Error())
		log.Fatal(err)
	}
	for _, rdsId := range rdsIDs {
		if opt.ToDelete {
			err := rdsClient.DeleteWhitelist(rdsId, opt.WhiteListName)
			if err != nil {
				log.Fatalf("Failed to delete whitelist %s under rdsid %s due to %v", opt.WhiteListName, rdsId, err)
			}
			log.Printf("Removed whitelist %s from rds %s\n", opt.WhiteListName, rdsId)
		} else {
			podExternalIP, err := utils.ExternalIP()
			if err != nil {
				msg := fmt.Sprintf("Failed to get pod's IP due to %v", err)
				msgLog.WriteString(msg)
				log.Fatalf(msg)
			}

			err = rdsClient.CreateWhitelist(rdsId, podExternalIP, opt.WhiteListName)
			if err != nil {
				msg := fmt.Sprintf("Failed to create whitelist %s under rdsid %s due to %v", opt.WhiteListName, rdsId, err)
				msgLog.WriteString(msg)
				log.Fatal(msg)
			}
			log.Printf("Created whitelist %s to rds %s\n", opt.WhiteListName, rdsId)
		}
	}
}

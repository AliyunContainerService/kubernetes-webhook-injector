package main

import (
	"github.com/AliyunContainerService/kubernetes-webhook-injector/pkg/openapi"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins/utils"
	"github.com/jessevdk/go-flags"
	"log"
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

var (
	opt       Options
	rdsClient *openapi.RdsWhitelistOperator
	rdsIDs    []string
)

func init() {
	_, err := flags.Parse(&opt)
	if err != nil {
		log.Fatal(err)
	}
	rdsIDs = strings.Split(opt.RdsIDs, ",")
	opt.WhiteListName = openapi.RefactorRdsWhitelistName(opt.WhiteListName)

	authInfo := &openapi.AKInfo{
		AccessKeyId:     opt.AccessKeyID,
		AccessKeySecret: opt.AccessKeySecret,
		SecurityToken:   opt.StsToken,
	}

	rdsClient, err = openapi.GetRdsWhitelistOperator(authInfo)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	for _, rdsId := range rdsIDs {
		if opt.ToDelete {
			log.Printf("Deleting whitelist %s from rds %s\n", opt.WhiteListName, rdsId)
			err := rdsClient.DeleteWhitelist(rdsId, opt.WhiteListName)
			if err != nil {
				log.Fatalf("Failed to delete whitelist %s under rdsid %s due to %v", opt.WhiteListName, rdsId, err)
			}
			log.Println("Deleted")

		} else {
			log.Printf("Creating whitelist %s to rds %s\n", opt.WhiteListName, rdsId)
			podExternalIP, err := utils.ExternalIP()
			if err != nil {
				log.Fatalf("Failed to get pod's IP due to %v", err)
			}

			err = rdsClient.CreateWhitelist(rdsId, podExternalIP, opt.WhiteListName)
			if err != nil {
				log.Fatalf("Failed to create whitelist %s under rdsid %s due to %v", opt.WhiteListName, rdsId, err)
			}
			log.Println("Created")

		}
	}
}

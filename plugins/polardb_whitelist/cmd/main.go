package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/AliyunContainerService/kubernetes-webhook-injector/pkg/openapi"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins/utils"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	PolarDBIDs      string `long:"polardb_id" required:"true"`
	RegionId        string `long:"region_id" required:"true"`
	WhiteListName   string `long:"white_list_name" required:"true"`
	AccessKeyID     string `long:"access_key_id" required:"true"`
	AccessKeySecret string `long:"access_key_secret" required:"true"`
	StsToken        string `long:"sts_token"`
	ToDelete        bool   `long:"delete"`
	IntranetAccess  bool   `long:"intranet_access"`
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
	polardbIDs := strings.Split(opt.PolarDBIDs, ",")

	openapi.InitClient(opt.IntranetAccess)

	authInfo := &openapi.AKInfo{
		AccessKeyId:     opt.AccessKeyID,
		AccessKeySecret: opt.AccessKeySecret,
		SecurityToken:   opt.StsToken,
	}

	polardbClient, err := openapi.GetPolarDBWhiteListOperator(authInfo)
	if err != nil {
		msgLog.WriteString(err.Error())
		log.Fatal(err)
	}

	podExternalIP, err := utils.ExternalIP()
	if err != nil {
		msg := fmt.Sprintf("Failed to get pod's IP due to %v", err)
		msgLog.WriteString(msg)
		log.Fatalf(msg)
	}

	for _, polardbID := range polardbIDs {
		if opt.ToDelete {
			err := polardbClient.DeleteWhitelist(polardbID, opt.WhiteListName, podExternalIP)
			if err != nil {
				log.Fatalf("Failed to delete whitelist %s under polardbId %s due to %v", opt.WhiteListName, polardbID, err)
			}
			log.Printf("Removed whitelist %s from polardb %s\n", opt.WhiteListName, polardbID)
		} else {
			err := polardbClient.CreateWhitelist(polardbID, podExternalIP, opt.WhiteListName)
			if err != nil {
				msg := fmt.Sprintf("Failed to create whitelist %s under polardbId %s due to %v", opt.WhiteListName, polardbID, err)
				msgLog.WriteString(msg)
				log.Fatal(msg)
			}
			log.Printf("Created whitelist %s to polardb %s\n", opt.WhiteListName, polardbID)
		}
	}
}

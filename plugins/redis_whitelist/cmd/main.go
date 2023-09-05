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
	RedisIDs        string `long:"redis_id" required:"true"`
	RegionId        string `long:"region_id" required:"true"`
	WhiteListName   string `long:"white_list_name" required:"true"`
	AccessKeyID     string `long:"access_key_id" required:"true"`
	AccessKeySecret string `long:"access_key_secret" required:"true"`
	StsToken        string `long:"sts_token"`
	ToDelete        bool   `long:"delete"`
	IntranetAccess  bool   `long:"intranet_access"`
}

var (
	opt    Options
	rdsIDs []string
)

const terminationLog = "/dev/termination-log"

func init() {

}

func main() {
	terminationLog, err := os.Create(terminationLog)
	if err != nil {
		log.Fatal(err)
	}

	defer terminationLog.Close()
	defer terminationLog.Sync()

	_, err = flags.Parse(&opt)
	if err != nil {
		terminationLog.WriteString(err.Error())
		log.Fatal(err)
	}

	redisIDs := strings.Split(opt.RedisIDs, ",")

	openapi.InitClient(opt.IntranetAccess)

	authInfo := &openapi.AKInfo{
		AccessKeyId:     opt.AccessKeyID,
		AccessKeySecret: opt.AccessKeySecret,
		SecurityToken:   opt.StsToken,
	}

	redisClient, err := openapi.GetRedisWhiteListOperator(authInfo)
	if err != nil {
		terminationLog.WriteString(err.Error())
		log.Fatal(err)
	}

	podExternalIP, err := utils.ExternalIP()
	if err != nil {
		msg := fmt.Sprintf("Failed to get pod's IP due to %v", err)
		terminationLog.WriteString(msg)
		log.Fatalf(msg)
	}

	for _, redisId := range redisIDs {
		if opt.ToDelete {
			err := redisClient.DeleteWhitelist(redisId, opt.WhiteListName, podExternalIP)
			if err != nil {
				log.Fatalf("Failed to delete whitelist %s under rdsid %s due to %v", opt.WhiteListName, redisId, err)
			}
			log.Printf("Removed whitelist %s from rds %s\n", opt.WhiteListName, redisId)
		} else {
			err := redisClient.CreateOrAppendWhitelist(redisId, podExternalIP, opt.WhiteListName)
			if err != nil {
				msg := fmt.Sprintf("Failed to create whitelist %s under rdsid %s due to %v", opt.WhiteListName, redisId, err)
				terminationLog.WriteString(msg)
				log.Fatal(msg)
			}
			log.Printf("Created whitelist %s to rds %s\n", opt.WhiteListName, redisId)
		}
	}
}

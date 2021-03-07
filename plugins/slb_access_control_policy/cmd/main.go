package main

import (
	"fmt"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/pkg/openapi"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins/utils"
	"github.com/jessevdk/go-flags"
	"log"
	"os"
)

type Options struct {
	RegionId              string `long:"region_id" required:"true"`
	AccessControlPolicyID string `long:"access_control_policy_id" required:"true"`
	AccessKeyID           string `long:"access_key_id" required:"true"`
	AccessKeySecret       string `long:"access_key_secret" required:"true"`
	StsToken              string `long:"sts_token"`
	ToDelete              bool   `long:"delete"`
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

	authInfo := &openapi.AKInfo{
		AccessKeyId:     opt.AccessKeyID,
		AccessKeySecret: opt.AccessKeySecret,
		SecurityToken:   opt.StsToken,
	}

	slbClient, err := openapi.GetSLBAccessControlPolicyOperator(authInfo)
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

	if opt.ToDelete {
		err := slbClient.DeleteEntryFromAccessControlPolicy(opt.AccessControlPolicyID, podExternalIP)
		if err != nil {
			log.Fatalf("Failed to delete entry from accessControlPolicy %s, due to %v", opt.AccessControlPolicyID, err)
		}
		log.Printf("Removed entry %s from accessControlPolicy %s\n", podExternalIP, opt.AccessControlPolicyID)
	} else {
		err := slbClient.AddEntryToAccessControlList(opt.AccessControlPolicyID, podExternalIP)
		if err != nil {
			msg := fmt.Sprintf("Failed to add %s to slb accessControlPolicy %s due to %v", podExternalIP, opt.AccessControlPolicyID, err)
			msgLog.WriteString(msg)
			log.Fatal(msg)
		}
		log.Printf("Add Entry %s to slb accessControlPolicy %s\n", podExternalIP, opt.AccessControlPolicyID)
	}
}

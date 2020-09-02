package rds_whitelist

import (
	"fmt"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/pkg/k8s"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/pkg/openapi"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins/utils"
	"k8s.io/api/admission/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	log "k8s.io/klog"
	"os"
	"strings"
)

const (
	PluginName = "RDSWhiteListPlugin"
	LabelRdsID = "ack.aliyun.com/rds_id"

	InitContainerName    = "rds-plugin"
	AddWlstCommandTemplt = "/root/rds-whitelist-plugin --region_id %s --rds_id %s --white_list_name %s --access_key_id %s --access_key_secret %s --sts_token %s"
	IMAGE_ENV            = "RDS_PLUGIN_IMAGE"
)

func init() {
	if image, ok := os.LookupEnv(IMAGE_ENV); ok {
		InitContainerImage = image
	}
}

var (
	matchLabels = map[string]string{
		LabelRdsID: "*",
	}

	cleaned = make(map[string]bool) //key: white list name

	InitContainerImage string
)

type rdsWhiteListPlugin struct {
	//authInfo openapi.AKInfo
}

func (r *rdsWhiteListPlugin) Name() string {
	return PluginName
}

func (r *rdsWhiteListPlugin) MatchAnnotations(podAnnots map[string]string) bool {
	for key, val := range matchLabels {
		if podV, ok := podAnnots[key]; ok {
			if val == "*" {
				continue
			}
			if podV != val {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func (r *rdsWhiteListPlugin) Patch(pod *apiv1.Pod, operation v1beta1.Operation) []utils.PatchOperation {
	var opPatches []utils.PatchOperation
	switch operation {
	case v1beta1.Create:
		for _, c := range pod.Spec.InitContainers {
			if c.Name == InitContainerName {
				break
			}
		}

		patch := r.patchInitContainer(pod)
		opPatches = append(opPatches, patch)
	case v1beta1.Delete:
		r.cleanUp(pod)
	}
	return opPatches
}

func (r *rdsWhiteListPlugin) patchInitContainer(pod *apiv1.Pod) utils.PatchOperation {
	authInfo, err := openapi.GetAuthInfo()
	if err != nil {
		log.Fatalf("Failed to authenticate OpenAPI client due to %v", err)
	}

	regionId := openapi.RegionID
	rdsId := pod.Annotations[LabelRdsID]
	akId := authInfo.AccessKeyId
	akSecrt := authInfo.AccessKeySecret
	stsToken := authInfo.SecurityToken

	con := apiv1.Container{
		Image:           InitContainerImage,
		Name:            InitContainerName,
		ImagePullPolicy: apiv1.PullAlways,
	}

	con.Command = strings.Split(
		fmt.Sprintf(AddWlstCommandTemplt, regionId, rdsId, "$(POD_NAMESPACE)_$(POD_NAME)", akId, akSecrt, stsToken), " ")

	con.Env = []apiv1.EnvVar{
		{Name: "POD_NAME", ValueFrom: &apiv1.EnvVarSource{FieldRef: &apiv1.ObjectFieldSelector{FieldPath: "metadata.name"}}},
		{Name: "POD_NAMESPACE", ValueFrom: &apiv1.EnvVarSource{FieldRef: &apiv1.ObjectFieldSelector{FieldPath: "metadata.namespace"}}},
		{Name: "REGION_ID", Value: regionId},
	}

	patch := utils.PatchOperation{
		Op:    "add",
		Path:  "/spec/initContainers",
		Value: []apiv1.Container{con},
	}
	return patch
}

func (r *rdsWhiteListPlugin) cleanUp(pod *apiv1.Pod) error {
	authInfo, err := openapi.GetAuthInfo()
	if err != nil {
		log.Fatalf("Failed to authenticate OpenAPI client due to %v", err)
	}

	rdsClient, err := openapi.GetRdsWhitelistOperator(authInfo)
	if err != nil {
		log.Fatal(err)
	}
	whitelistName := openapi.RefactorRdsWhitelistName(pod.Namespace + "_" + pod.Name)
	//删除Pod时，webhook会触发3次
	if cleaned[whitelistName] {
		return nil
	}
	cleaned[whitelistName] = true

	go func() {
		rdsIDs := strings.Split(pod.Annotations[LabelRdsID], ",")
		for _, rdsId := range rdsIDs {

			err := rdsClient.DeleteWhitelist(rdsId, whitelistName)
			if err != nil {
				msg := fmt.Sprintf("Failed to delete whitelist %s under rds %s due to %v", whitelistName, rdsId, err)
				log.Error(msg)
				k8s.GetEventor().SendPodEvent(pod, apiv1.EventTypeWarning, "Deleting", msg)
				return
			}
			msg := fmt.Sprintf("removed whitelist %s from rds %s", whitelistName, rdsId)
			log.Infof(msg)
			k8s.GetEventor().SendPodEvent(pod, apiv1.EventTypeNormal, "Deleting", msg)
		}
	}()

	return nil
}

func NewRdsPlugin() *rdsWhiteListPlugin {
	return &rdsWhiteListPlugin{}
}

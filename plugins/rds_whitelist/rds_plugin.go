package rds_whitelist

import (
	"fmt"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/pkg/k8s"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/pkg/openapi"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins/utils"
	admissionv1 "k8s.io/api/admission/v1"
	apiv1 "k8s.io/api/core/v1"
	log "k8s.io/klog"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	PluginName         = "RDSWhiteListPlugin"
	LabelRdsID         = "ack.aliyun.com/rds_id"
	LabelWhiteListName = "ack.aliyun.com/white_list_name"

	InitContainerName    = "rds-plugin"
	AddWlstCommandTemplt = "/root/rds-whitelist-plugin --region_id %s --rds_id %s --white_list_name %s --access_key_id %s --access_key_secret %s --sts_token %s %s"
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

	//cleaned = make(map[string]bool) //key: white list name
	checked = make(map[string]bool) //key: pod name 防止Pod中 init 容器被反复检查状态

	InitContainerImage string

	lock = &sync.Mutex{}
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

func (r *rdsWhiteListPlugin) Patch(pod *apiv1.Pod, operation admissionv1.Operation, option *utils.PluginOption) []utils.PatchOperation {
	var opPatches []utils.PatchOperation
	switch operation {
	case admissionv1.Create:
		for _, c := range pod.Spec.InitContainers {
			if c.Name == InitContainerName {
				break
			}
		}

		patch := r.patchInitContainer(pod, option)
		opPatches = append(opPatches, patch)
		go func() {
			// 获取相同命名空间下，Generate名相同，并且有rds-plugin为名字的init容器
			chPod := k8s.GetPodsByPluginNameCh(pod.Namespace, pod.GenerateName, InitContainerName)
			for pod := range chPod {
				if _, ok := checked[pod.Name]; ok {
					continue
				}
				lock.Lock()
				checked[pod.Name] = true
				lock.Unlock()

				ch := make(chan k8s.ContainerExitStatus, 1)
				go k8s.WatchInitContainerStatus2(pod, InitContainerName, ch)

				select {
				case status := <-ch:
					if !status.Success {
						log.Error(status.Message)
						if status.Pod != nil {
							rdsId := pod.Annotations[LabelRdsID]
							k8s.SendPodEvent(status.Pod, apiv1.EventTypeWarning, "Created", status.Message+" rdsId: "+rdsId)
						}
					} else {
						log.Info(status.Message)
					}
				case <-time.After(70 * time.Second):
					log.Error("Plugin execution times out, please check pod status manually")
				}
			}
		}()
	case admissionv1.Delete:
		r.cleanUp(pod)
	}
	return opPatches
}

func (r *rdsWhiteListPlugin) patchInitContainer(pod *apiv1.Pod, option *utils.PluginOption) utils.PatchOperation {
	authInfo, err := openapi.GetAuthInfo()
	if err != nil {
		log.Fatalf("Failed to authenticate OpenAPI client due to %v", err)
	}

	regionId := openapi.RegionID
	rdsId := pod.Annotations[LabelRdsID]
	whiteListName := pod.Annotations[LabelWhiteListName]

	akId := authInfo.AccessKeyId
	akSecrt := authInfo.AccessKeySecret
	stsToken := authInfo.SecurityToken

	access := ""
	if option.IntranetAccess {
		access = "--intranet_access"
	}

	con := apiv1.Container{
		Image:           InitContainerImage,
		Name:            InitContainerName,
		ImagePullPolicy: apiv1.PullAlways,
	}

	con.Command = strings.Split(
		fmt.Sprintf(AddWlstCommandTemplt, regionId, rdsId, whiteListName, akId, akSecrt, stsToken, access), " ")

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

	whitelistName := pod.Annotations[LabelWhiteListName]
	//删除Pod时，webhook会触发3次
	//if cleaned[whitelistName] {
	//	return nil
	//}
	//cleaned[whitelistName] = true

	go func(whiteList string) {
		rdsIDs := strings.Split(pod.Annotations[LabelRdsID], ",")
		for _, rdsId := range rdsIDs {
			err := rdsClient.DeleteWhitelist(rdsId, whiteList, pod.Status.PodIP)
			if err != nil {
				msg := fmt.Sprintf("Failed to delete %v from whitelist %s under rds %s. %s. %s",
					pod.Status.PodIP, whiteList, rdsId, openapi.ParseErrorMessage(err.Error()).ErrorCode, openapi.ParseErrorMessage(err.Error()).Message)
				log.Error(msg)
				k8s.SendPodEvent(pod, apiv1.EventTypeWarning, "Deleting", msg)
				return
			}
			msg := fmt.Sprintf("removed %v from whitelist %s in rds %s", pod.Status.PodIP, whitelistName, rdsId)
			log.Infof(msg)
			k8s.SendPodEvent(pod, apiv1.EventTypeNormal, "Deleting", msg)
		}
	}(whitelistName)

	return nil
}

func NewRdsPlugin() *rdsWhiteListPlugin {
	return &rdsWhiteListPlugin{}
}

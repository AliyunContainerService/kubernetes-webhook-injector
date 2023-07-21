package redis_whitelist

import (
	"fmt"
	"k8s.io/klog"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/AliyunContainerService/kubernetes-webhook-injector/pkg/k8s"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/pkg/openapi"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins/utils"
	admissionv1 "k8s.io/api/admission/v1"
	apiv1 "k8s.io/api/core/v1"
)

const (
	LabelRedisID         = "ack.aliyun.com/redis_id"
	PluginName           = "RedisWhiteListPlugin"
	InitContainerName    = "redis-plugin"
	LabelWhiteListName   = "ack.aliyun.com/redis_white_list_name"
	IMAGE_ENV            = "REDIS_PLUGIN_IMAGE"
	AddWlstCommandTemplt = "/root/redis-whitelist-plugin --region_id %s --redis_id %s --white_list_name %s --access_key_id %s --access_key_secret %s --sts_token %s"
)

var (
	matchLabels = map[string]string{
		LabelRedisID: "*",
	}

	InitContainerImage string

	checked = make(map[string]bool)

	lock = &sync.Mutex{}
)

type redisWhiteListPlugin struct{}

func init() {
	if image, ok := os.LookupEnv(IMAGE_ENV); ok {
		InitContainerImage = image
	}
}

func NewRedisPlugin() *redisWhiteListPlugin {
	return &redisWhiteListPlugin{}
}

func (rp *redisWhiteListPlugin) Name() string {
	return PluginName
}

func (rp *redisWhiteListPlugin) MatchAnnotations(podAnnots map[string]string) bool {
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

func (rp *redisWhiteListPlugin) Patch(pod *apiv1.Pod, operation admissionv1.Operation) []utils.PatchOperation {
	var opPatches []utils.PatchOperation
	switch operation {
	case admissionv1.Create:
		for _, container := range pod.Spec.InitContainers {
			if container.Name == InitContainerName {
				break
			}
		}

		patch := rp.patchInitContainer(pod)
		opPatches = append(opPatches, patch)
		go func() {
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
						klog.Error(status.Message)
						if status.Pod != nil {
							redisId := pod.Annotations[LabelRedisID]
							k8s.SendPodEvent(status.Pod, apiv1.EventTypeWarning, "Created", status.Message+" redisId: "+redisId)
						}
					} else {
						klog.Info(status.Message)
					}
				case <-time.After(70 * time.Second):
					klog.Error("Plugin execution times out, please check pod status manually")
				}
			}
		}()

	case admissionv1.Delete:
		rp.cleanUp(pod)
	}
	return opPatches
}

func (rp *redisWhiteListPlugin) patchInitContainer(pod *apiv1.Pod) utils.PatchOperation {
	authInfo, err := openapi.GetAuthInfo()
	if err != nil {
		log.Fatalf("Failed to authenticate OpenAPI client due to %v", err)
	}

	regionId := openapi.RegionID
	redisId := pod.Annotations[LabelRedisID]
	whiteListName := pod.Annotations[LabelWhiteListName]

	akId := authInfo.AccessKeyId
	akSecrt := authInfo.AccessKeySecret
	stsToken := authInfo.SecurityToken

	con := apiv1.Container{
		Image:           InitContainerImage,
		Name:            InitContainerName,
		ImagePullPolicy: apiv1.PullAlways,
	}

	con.Command = strings.Split(
		fmt.Sprintf(AddWlstCommandTemplt, regionId, redisId, whiteListName, akId, akSecrt, stsToken), " ")

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

func (rp *redisWhiteListPlugin) cleanUp(pod *apiv1.Pod) error {
	authInfo, err := openapi.GetAuthInfo()
	if err != nil {
		log.Fatalf("Failed to authenticate OpenAPI client due to %v", err)
	}

	redisClient, err := openapi.GetRedisWhiteListOperator(authInfo)
	if err != nil {
		log.Fatal(err)
	}

	whitelistName := pod.Annotations[LabelWhiteListName]

	go func() {
		redisIDs := strings.Split(pod.Annotations[LabelRedisID], ",")
		for _, redisId := range redisIDs {
			err := redisClient.DeleteWhitelist(redisId, whitelistName, pod.Status.PodIP)
			if err != nil {
				msg := fmt.Sprintf("Failed to delete %v from whitelist %s under redis %s %s",
					pod.Status.PodIP, whitelistName, redisId, openapi.ParseErrorMessage(err.Error()).Message)
				klog.Error(msg)
				k8s.SendPodEvent(pod, apiv1.EventTypeWarning, "Deleting", msg)
				return
			}
			msg := fmt.Sprintf("removed %v from whitelist %s in redis %s", pod.Status.PodIP, whitelistName, redisId)
			klog.Infof(msg)
			k8s.SendPodEvent(pod, apiv1.EventTypeNormal, "Deleting", msg)
		}
	}()

	return nil
}

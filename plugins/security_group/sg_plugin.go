package security_group

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
	"sync"
	"time"
)

const (
	PluginName = "SecurityGroupPlugin"
	LabelSgID  = "ack.aliyun.com/security_group_id"

	AddSgCommandTemplt = "/root/security-group-plugin --region_id %s --security_group_id %s --permission_id %s --access_key_id %s --access_key_secret %s --sts_token %s"
	InitContainerName  = "sg-plugin"
	IMAGE_ENV          = "SG_PLUGIN_IMAGE"
)

func init() {
	if image, ok := os.LookupEnv(IMAGE_ENV); ok {
		InitContainerImage = image
	}
}

var (
	matchLabels = map[string]string{
		LabelSgID: "*",
	}

	cleaned = make(map[string]bool) //key: sg permission desc
	checked = make(map[string]bool) //key: pod name

	InitContainerImage string
	lock               = &sync.Mutex{}
)

type SecurityGroupPlugin struct {
}

func (s *SecurityGroupPlugin) cleanUp(pod *apiv1.Pod) error {
	//TODO 这部分逻辑可以移到 utils 里

	// 因为sts会更新，所以每次都现取
	authInfo, err := openapi.GetAuthInfo()
	if err != nil {
		log.Fatalf("Failed to authenticate OpenAPI client due to %v", err)
	}

	sgClient, err := openapi.GetSecurityGroupOperator(authInfo)
	if err != nil {
		log.Fatal(err)
	}
	pDesc := pod.Namespace + ":" + pod.Name

	//删除Pod时，webhook会触发3次
	if cleaned[pDesc] {
		return nil
	}
	cleaned[pDesc] = true

	go func() {
		sgIDs := strings.Split(pod.Annotations[LabelSgID], ",")
		for _, sgID := range sgIDs {
			p, err := sgClient.FindPermission(sgID, pDesc)
			if err != nil {
				openapiMsg := openapi.ParseErrorMessage(err.Error()).Message
				eventMsg := fmt.Sprintf("Failed to find security group with ID %s %s", sgID, openapiMsg)
				log.Infof(eventMsg)
				k8s.SendPodEvent(pod, apiv1.EventTypeWarning, "Deleting", eventMsg)
				continue
			}
			//如果有，先删除
			if p == nil {
				msg := fmt.Sprintf("Cannot find permission %s of sg %s in region %s\n", pDesc, sgID, openapi.RegionID)
				log.Infof(msg)
				k8s.SendPodEvent(pod, apiv1.EventTypeWarning, "Deleting", fmt.Sprintf(msg))
			} else {
				err = sgClient.DeletePermission(sgID, p)
				if err != nil {
					msg := fmt.Sprintf("failed to remove permission %s of sg %s in region %s %s",
						pDesc, sgID, openapi.RegionID, openapi.ParseErrorMessage(err.Error()).Message)
					log.Error(msg)
					k8s.SendPodEvent(pod, apiv1.EventTypeNormal, "Deleting", msg)
				}
				msg := fmt.Sprintf("Removed permission %s of sg %s in region %s", pDesc, sgID, openapi.RegionID)
				log.Infof(msg)
				k8s.SendPodEvent(pod, apiv1.EventTypeNormal, "Deleting", msg)
			}
		}
	}()
	return nil
}

func (s *SecurityGroupPlugin) Name() string {
	return PluginName
}

func (s *SecurityGroupPlugin) MatchAnnotations(podAnnots map[string]string) bool {
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
func (s *SecurityGroupPlugin) Patch(pod *apiv1.Pod, operation v1beta1.Operation) []utils.PatchOperation {
	var patches []utils.PatchOperation

	switch operation {
	case v1beta1.Create:
		//如果已经有sg initContainer：保证幂等
		for _, c := range pod.Spec.InitContainers {
			if c.Name == InitContainerName {
				break
			}
		}
		patch := s.patchInitContainer(pod)
		patches = append(patches, patch)
		go func() {
			//todo 查找同命名空间下所有相同 generationName 的Pods，
			//     for each pod：
			//          如果不在checked map中，将pod名加入进去, 如果已有，直接跳过
			//          为每个pod启动一个 goroutine，等待10秒，读取 plugin container的状态，如果异常，发Event

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
							sgId := pod.Annotations[LabelSgID]
							k8s.SendPodEvent(status.Pod, apiv1.EventTypeWarning, "Created", status.Message+" gs Id: "+sgId)
						}
					} else {
						log.Info(status.Message)
					}
				case <-time.After(70 * time.Second):
					log.Error("Plugin execution times out, please check pod status manually")
				}
			}
		}()
	case v1beta1.Delete:
		s.cleanUp(pod)

	case v1beta1.Update:
		log.Infof("Pod %s is updated", pod.Name)
	}

	return patches
}

func (s *SecurityGroupPlugin) patchInitContainer(pod *apiv1.Pod) utils.PatchOperation {
	authInfo, err := openapi.GetAuthInfo()
	if err != nil {
		log.Fatalf("Failed to authenticate OpenAPI client due to %v", err)
	}

	regionId := openapi.RegionID
	sgId := pod.Annotations[LabelSgID]
	akId := authInfo.AccessKeyId
	akSecrt := authInfo.AccessKeySecret
	stsToken := authInfo.SecurityToken

	con := apiv1.Container{
		Image:           InitContainerImage,
		Name:            InitContainerName,
		ImagePullPolicy: apiv1.PullAlways,
	}

	con.Command = strings.Split(fmt.Sprintf(AddSgCommandTemplt, regionId, sgId,
		"$(POD_NAMESPACE):$(POD_NAME)", akId, akSecrt, stsToken), " ")

	con.Env = []apiv1.EnvVar{
		{Name: "POD_NAME", ValueFrom: &apiv1.EnvVarSource{FieldRef: &apiv1.ObjectFieldSelector{FieldPath: "metadata.name"}}},
		{Name: "POD_NAMESPACE", ValueFrom: &apiv1.EnvVarSource{FieldRef: &apiv1.ObjectFieldSelector{FieldPath: "metadata.namespace"}}},
		{Name: "REGION_ID", Value: regionId},
	}

	return utils.PatchOperation{
		Op:    "add",
		Path:  "/spec/initContainers",
		Value: []apiv1.Container{con},
	}
}

func NewSgPlugin() *SecurityGroupPlugin {
	return &SecurityGroupPlugin{}
}

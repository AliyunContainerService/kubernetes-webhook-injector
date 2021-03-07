package slb_access_control_policy

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
	"k8s.io/api/admission/v1beta1"
	apiv1 "k8s.io/api/core/v1"
)

const (
	PluginName                 = "SLBWhiteListPlugin"
	LabelAccessControlPolicyID = "ack.aliyun.com/access_control_policy_id"

	InitContainerName    = "slb-plugin"
	AddWlstCommandTemplt = "/root/slb-access-control-plugin --region_id %s --access_control_policy_id %s --access_key_id %s --access_key_secret %s --sts_token %s"
	IMAGE_ENV            = "SLB_PLUGIN_IMAGE"
)

var (
	matchLabels = map[string]string{
		LabelAccessControlPolicyID: "*",
	}

	//cleaned = make(map[string]bool) //key: white list name
	checked = make(map[string]bool) //key: pod name 防止Pod中 init 容器被反复检查状态

	InitContainerImage string

	lock = &sync.Mutex{}
)

func init() {
	if image, ok := os.LookupEnv(IMAGE_ENV); ok {
		InitContainerImage = image
	}
}

type SLBPlugin struct{}

func NewSLBPlugin() *SLBPlugin {
	return &SLBPlugin{}
}

func (sp *SLBPlugin) Name() string {
	return PluginName
}

func (sp *SLBPlugin) MatchAnnotations(podAnnots map[string]string) bool {

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

func (sp *SLBPlugin) Patch(pod *apiv1.Pod, operation v1beta1.Operation) []utils.PatchOperation {
	var opPatches []utils.PatchOperation
	switch operation {
	case v1beta1.Create:
		for _, c := range pod.Spec.InitContainers {
			if c.Name == InitContainerName {
				break
			}
		}

		patch := sp.patchInitContainer(pod)
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
							accessControlPolicy := pod.Annotations[LabelAccessControlPolicyID]
							k8s.SendPodEvent(status.Pod, apiv1.EventTypeWarning, "Created", status.Message+" access control policy : "+accessControlPolicy)
						}
					} else {
						klog.Info(status.Message)
					}
				case <-time.After(70 * time.Second):
					klog.Error("Plugin execution times out, please check pod status manually")
				}
			}
		}()
	case v1beta1.Delete:
		sp.cleanUp(pod)
	}
	return opPatches
}

func (sp *SLBPlugin) patchInitContainer(pod *apiv1.Pod) utils.PatchOperation {
	authInfo, err := openapi.GetAuthInfo()
	if err != nil {
		log.Fatalf("Failed to authenticate OpenAPI client due to %v", err)
	}

	regionId := openapi.RegionID
	slbAccessControlPolicy := pod.Annotations[LabelAccessControlPolicyID]

	akId := authInfo.AccessKeyId
	akSecrt := authInfo.AccessKeySecret
	stsToken := authInfo.SecurityToken

	con := apiv1.Container{
		Image:           InitContainerImage,
		Name:            InitContainerName,
		ImagePullPolicy: apiv1.PullAlways,
	}

	con.Command = strings.Split(
		fmt.Sprintf(AddWlstCommandTemplt, regionId, slbAccessControlPolicy, akId, akSecrt, stsToken), " ")

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

func (sp *SLBPlugin) cleanUp(pod *apiv1.Pod) error {
	authInfo, err := openapi.GetAuthInfo()
	if err != nil {
		log.Fatalf("Failed to authenticate OpenAPI client due to %v", err)
	}

	slbClient, err := openapi.GetSLBAccessControlPolicyOperator(authInfo)
	if err != nil {
		log.Fatal(err)
	}

	accessControlPolicyID := pod.Annotations[LabelAccessControlPolicyID]
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err := slbClient.DeleteEntryFromAccessControlPolicy(accessControlPolicyID, pod.Status.PodIP)
		if err != nil {
			msg := fmt.Sprintf("Failed to delete %v from slb accessControlPolicy %s: %v",
				pod.Status.PodIP, accessControlPolicyID, openapi.ParseErrorMessage(err.Error()).Message)
			klog.Error(msg)
			k8s.SendPodEvent(pod, apiv1.EventTypeWarning, "Deleting", msg)
			return
		}
		msg := fmt.Sprintf("removed %v from slb accessControlPolicy %s.", pod.Status.PodIP, accessControlPolicyID)
		klog.Infof(msg)
		k8s.SendPodEvent(pod, apiv1.EventTypeNormal, "Deleting", msg)
	}()

	return nil
}

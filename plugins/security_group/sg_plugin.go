package security_group

import (
	"fmt"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/pkg/openapi"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins/utils"
	"k8s.io/api/admission/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	log "k8s.io/klog"
	"os"
	"strings"
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

	InitContainerImage string
)

type SecurityGroupPlugin struct {
	//initImage string
	//authInfo openapi.AKInfo
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
		// 打印错误消息，发k8s event
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
			log.Infof("Deleting permission %s of sg %s in region %s\n", pDesc, sgID, openapi.RegionID)
			p, err := sgClient.FindPermission(sgID, pDesc)
			if err != nil {
				log.Fatal(err)
			}
			//如果有，先删除
			if p == nil {
				log.Infof("WARNING: Cannot find permission %s of sg %s in region %s\n", pDesc, sgID, openapi.RegionID)
			} else {
				err = sgClient.DeletePermission(sgID, p)
				if err != nil {
					log.Fatal(err)
				}
				log.Infof("Permission deleted")
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
	case v1beta1.Delete:
		s.cleanUp(pod)
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

//jobsClient := utils.ClientSet.BatchV1().Jobs(pod.Namespace)
//	job := &batchv1.Job{
//		ObjectMeta: metav1.ObjectMeta{
//			Namespace: pod.Namespace,
//			Name:      pod.Name + "-sg-" + "cleaner",
//		},
//		Spec: batchv1.JobSpec{
//			Template: apiv1.PodTemplateSpec{
//				Spec: apiv1.PodSpec{
//					RestartPolicy: apiv1.RestartPolicyOnFailure,
//					Containers: []apiv1.Container{
//						{
//							Name:            InitContainerName,
//							Image:           s.initImage,
//							ImagePullPolicy: apiv1.PullAlways,
//							//Command:         append(s.initCmd, "--delete"),
//							Env: []apiv1.EnvVar{
//								{Name: "ACCESS_KEY_ID", Value: s.authInfo.AccessKeyId},
//								{Name: "ACCESS_KEY_SECRET", Value: s.authInfo.AccessKeySecret},
//								{Name: "POD_NAME", Value: pod.Name},
//								{Name: "POD_NAMESPACE", ValueFrom: &apiv1.EnvVarSource{FieldRef: &apiv1.ObjectFieldSelector{FieldPath: "metadata.namespace"}}},
//								{Name: "REGION_ID", Value: pod.Annotations[LabelRegion]},
//								{Name: "SECURITY_GROUP_ID", Value: pod.Annotations[LabelSgID]},
//							},
//						},
//					},
//				},
//			},
//		},
//	}

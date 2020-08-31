package plugins

import (
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins/utils"
	"k8s.io/api/admission/v1beta1"
	apiv1 "k8s.io/api/core/v1"
)

type Plugin interface {
	Name() string
	MatchAnnotations(map[string]string) bool
	Patch(*apiv1.Pod, v1beta1.Operation) []utils.PatchOperation
}

package plugins

import (
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins/utils"
	admissionv1 "k8s.io/api/admission/v1"
	apiv1 "k8s.io/api/core/v1"
)

type Plugin interface {
	Name() string
	MatchAnnotations(map[string]string) bool
	Patch(*apiv1.Pod, admissionv1.Operation, *utils.PluginOption) []utils.PatchOperation
}

package plugins

import (
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins/utils"
	apiv1 "k8s.io/api/core/v1"
)

type Plugin interface {
	Name() string
	MatchAnnotations(map[string]string) bool
	Patch(pod *apiv1.Pod) []utils.PatchOperation
}

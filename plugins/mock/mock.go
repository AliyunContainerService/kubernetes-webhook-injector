package mock

import (
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins/utils"
	"k8s.io/api/admission/v1beta1"
	apiv1 "k8s.io/api/core/v1"
)

// mock plugin implement
const (
	MockPluginName = "Mock"
)

type MockPlugin struct{}

func init() {
	//plugins.Register(NewMockPlugin())
}

func (mp *MockPlugin) Name() string {
	return MockPluginName
}

func (mp *MockPlugin) MatchAnnotations(map[string]string) bool {
	return false
}

func (mp *MockPlugin) Patch(pod *apiv1.Pod, operation v1beta1.Operation) []utils.PatchOperation {
	return nil
}

func NewMockPlugin() *MockPlugin {
	return &MockPlugin{}
}

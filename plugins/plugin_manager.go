package plugins

import (
	"encoding/json"
	"fmt"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/pkg/openapi"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins/mock"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins/rds_whitelist"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins/redis_whitelist"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins/security_group"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins/slb_access_control_policy"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins/utils"
	"k8s.io/api/admission/v1beta1"
	apiv1 "k8s.io/api/core/v1"
	log "k8s.io/klog"
)

var (
	pluginManagerSingleton *PluginManager
	OpenAPIAuthInfo        openapi.AKInfo
)

func init() {
	pluginManagerSingleton = &PluginManager{
		plugins: make(map[string]Plugin),
	}

	// register mock plugin
	pluginManagerSingleton.register(mock.NewMockPlugin())
	pluginManagerSingleton.register(security_group.NewSgPlugin())
	pluginManagerSingleton.register(rds_whitelist.NewRdsPlugin())
	pluginManagerSingleton.register(redis_whitelist.NewRedisPlugin())
	pluginManagerSingleton.register(slb_access_control_policy.NewSLBPlugin())
}

type PluginManager struct {
	plugins map[string]Plugin
}

// register plugin to manster
func (pm *PluginManager) register(plugin Plugin) (err error) {
	if name := plugin.Name(); name != "" {
		pm.plugins[name] = plugin
		return nil
	}

	return fmt.Errorf("plugin %v is invalid", plugin)
}

// handle patch pod operations
func (pm *PluginManager) HandlePatchPod(pod *apiv1.Pod, operation v1beta1.Operation) ([]byte, error) {
	patchOperations := make([]utils.PatchOperation, 0)
	for _, plugin := range pm.plugins {
		if plugin.MatchAnnotations(pod.Annotations) {
			//如果是创建，调用Patch为Pod注入init容器做各种操作
			singlePatchOperations := plugin.Patch(pod, operation)
			patchOperations = append(patchOperations, singlePatchOperations...)
		}
	}
	//originalContainers := pod.Spec.InitContainers
	if len(patchOperations) > 0 {
		//toPatch := mergePatchOperations(patchOperations)
		//toPatch
		patchBytes, err := json.Marshal(mergePatchOperations(patchOperations, pod.Spec.InitContainers))
		if err != nil {
			log.Warningf("Failed to marshal patch bytes by plugin skip,because of %v", err)
		} else {
			return patchBytes, nil
		}
	}

	// no match any one
	return nil, nil
}

func mergePatchOperations(operations []utils.PatchOperation, originalContainers []apiv1.Container) []utils.PatchOperation {
	mgdOps := make([]utils.PatchOperation, 0)
	containers := originalContainers
	for _, op := range operations {
		if op.Op == "add" && op.Path == "/spec/initContainers" {
			if l, ok := op.Value.([]apiv1.Container); ok {
				containers = append(containers, l...)
			}
		} else {
			mgdOps = append(mgdOps, op)
		}
	}
	mgdOps = append(mgdOps, utils.PatchOperation{Op: "add", Path: "/spec/initContainers", Value: containers})
	return mgdOps
}

// return singleton
func NewPluginManager() *PluginManager {
	return pluginManagerSingleton
}

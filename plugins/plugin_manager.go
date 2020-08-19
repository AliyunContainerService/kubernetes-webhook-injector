package plugins

import (
	"encoding/json"
	"fmt"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins/mock"
	apiv1 "k8s.io/api/core/v1"
	log "k8s.io/klog"
)

var (
	pluginManagerSingleton *PluginManager
)

func init() {
	pluginManagerSingleton = &PluginManager{
		plugins: make(map[string]Plugin),
	}

	// register mock plugin
	pluginManagerSingleton.register(mock.NewMockPlugin())
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
func (pm *PluginManager) HandlePatchPod(pod *apiv1.Pod) ([]byte, error) {
	for _, plugin := range pm.plugins {
		if plugin.MatchAnnotations(pod.Annotations) {
			patchOperations := plugin.Patch(pod)
			if len(patchOperations) == 0 {
				continue
			} else {
				patchBytes, err := json.Marshal(patchOperations)
				if err != nil {
					log.Warningf("Failed to marshal patch bytes by plugin %s and skip,because of %v", plugin.Name(), err)
					continue
				} else {
					return patchBytes, nil
				}
			}
		}
	}
	// no match any one
	return nil, nil
}

// return singleton
func NewPluginManager() *PluginManager {
	return pluginManagerSingleton
}

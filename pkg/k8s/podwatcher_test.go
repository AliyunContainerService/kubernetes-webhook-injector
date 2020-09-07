package k8s

import (
	"fmt"
	"os"
	"testing"
)

var (
	ns             = "mutating-inject"
	generationName = "inject-test-649b599897-"
)

func TestPodWatcher_WatchPlugin(t *testing.T) {
	if _, isLocalTest := os.LookupEnv("LocalTestEnv"); !isLocalTest {
		t.SkipNow()
	}
	InitClientSetOrDie("", "/Users/ruijzhan/.kube/config")

	ch := GetPodsByPluginNameCh(ns, generationName, "sg-plugin")
	for pod := range ch {
		fmt.Println(pod.Name)
	}
}
func TestWatchInitContainerStatus(t *testing.T) {
	if _, isLocalTest := os.LookupEnv("LocalTestEnv"); !isLocalTest {
		t.SkipNow()
	}

	ns := "mutating-inject"
	generationName := "inject-test-569444f459-"
	pluginName := "rds-plugin"

	InitClientSetOrDie("", "/Users/ruijzhan/.kube/config")
	ch := GetPodsByPluginNameCh(ns, generationName, pluginName)
	for pod := range ch {
		WatchInitContainerStatus(pod, pluginName)
	}
}

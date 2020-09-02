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

func ini() {
}

func TestPodWatcher_WatchPlugin(t *testing.T) {
	if _, isLocalTest := os.LookupEnv("LocalTestEnv"); !isLocalTest {
		t.SkipNow()
	}
	cs := GetClientSetOrDie("", "/Users/ruijzhan/.kube/config")

	ch := GetPodsByPluginNameCh(cs, ns, generationName, "sg-plugin")
	for pod := range ch {
		fmt.Println(pod.Name)
	}
}
func TestWatchInitContainerStatus(t *testing.T) {
	if _, isLocalTest := os.LookupEnv("LocalTestEnv"); !isLocalTest {
		t.SkipNow()
	}

	ns := "mutating-inject"
	generationName := "inject-test-649b599897-"
	pluginName := "sg-plugin"

	cs := GetClientSetOrDie("", "/Users/ruijzhan/.kube/config")
	evtor := NewEventor(cs)
	ch := GetPodsByPluginNameCh(cs, ns, generationName, pluginName)
	for pod := range ch {
		WatchInitContainerStatus(cs, pod, evtor)
	}
}

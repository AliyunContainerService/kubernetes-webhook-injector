package k8s

import (
	"context"
	"fmt"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/pkg/openapi"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	log "k8s.io/klog"
	"time"
)

type PodWatcher struct {
	ClientSet kubernetes.Interface
}

var (
//podWatcher *PodWatcher
)

func NewPodWatcher(client kubernetes.Interface) *PodWatcher {
	//client := GetClientSetOrDie(masterUrl, kubeConfigPath)

	return &PodWatcher{
		ClientSet: client,
	}
}

//Watch 给定命名空间下，相同generatedName的，具有pluginName名字的init container的Pod的创建操作，用通道返回
func GetPodsByPluginNameCh(clientSet kubernetes.Interface, namespace, generateName, pluginName string) <-chan *apiv1.Pod {
	ch := make(chan *apiv1.Pod)
	watcher, err := clientSet.CoreV1().Pods(namespace).Watch(
		context.Background(),
		metav1.ListOptions{},
	)
	if err != nil {
		log.Error(err)
		//todo 这里发event？
	}
	go func() {
		<-time.NewTimer(10 * time.Second).C //10 seconds to wait for pod actually been created in cluster
		watcher.Stop()
	}()

	go func() {
		for event := range watcher.ResultChan() {
			if event.Type == watch.Added {
				pod, ok := event.Object.(*apiv1.Pod)
				if !ok {
					log.Fatal("Unexpected type while watching pod events")
				}
				if pod.GenerateName == generateName {
					for _, initC := range pod.Spec.InitContainers {
						if initC.Name == pluginName {
							ch <- pod
						}
					}
				}
			}
		}
		close(ch)
	}()
	return ch
}

func GetPod(cs kubernetes.Interface, namespace, name string) (*apiv1.Pod, error) {
	return cs.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

//跟踪pod的init container的启动情况
func WatchInitContainerStatus(cs kubernetes.Interface, pod *apiv1.Pod, eventor *Eventor) {
	for i := 0; i < 3; i++ {
		time.Sleep(5 * time.Second)
		currentPod, err := GetPod(cs, pod.Namespace, pod.Name)
		if err != nil {
			log.Error(fmt.Sprintf("Cannot get pod %s from namespace %s", pod.Name, pod.Namespace))
			return
		}
		for _, status := range currentPod.Status.InitContainerStatuses {
			if status.LastTerminationState.Terminated != nil {
				if status.LastTerminationState.Terminated.ExitCode != 0 {
					msg := fmt.Sprintf("plugin %s failed in pod %s of namespace %s. %s",
						status.Name, pod.Name, pod.Namespace, openapi.ParseErrorMessage(status.LastTerminationState.Terminated.Message).Message)
					log.Error(msg)
					eventor.SendPodEvent(currentPod, apiv1.EventTypeWarning, "Created", msg)
					return
				}
			}

			if status.State.Terminated != nil {
				if status.State.Terminated.ExitCode == 0 {
					log.Infof("plugin %s executed successfully in pod %s of namespace %s",
						status.Name, pod.Name, pod.Namespace)
					return
				}
			}
		}
	}
}

//func WatchPod(clientSet kubernetes.Interface, pod *apiv1.Pod) {
//	watch, err := clientSet.CoreV1().Pods(pod.Namespace).Watch(context.Background(), metav1.ListOptions{})
//	if err != nil {
//		log.Fatal(err)
//	}
//	go func() {
//		for event := range watch.ResultChan() {
//			fmt.Printf("Type: %v\n", event.Type)
//			p, ok := event.Object.(*apiv1.Pod)
//			if !ok {
//				log.Fatal("unexpected type")
//			}
//			fmt.Println(p.Status.ContainerStatuses)
//			fmt.Println(p.Status.Phase)
//		}
//	}()
//	time.Sleep(60 * time.Second)
//}

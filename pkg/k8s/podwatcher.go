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
	//client := InitClientSetOrDie(masterUrl, kubeConfigPath)

	return &PodWatcher{
		ClientSet: client,
	}
}

//Watch 给定命名空间下，相同generatedName的，具有pluginName名字的init container的Pod的创建操作，用通道返回
func GetPodsByPluginNameCh(namespace, generateName, pluginName string) <-chan *apiv1.Pod {
	clientSet := GetClientSet()
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

func GetPod(namespace, name string) (*apiv1.Pod, error) {
	cs := GetClientSet()
	return cs.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
}

type ContainerExitStatus struct {
	Success bool
	Message string
	Pod     *apiv1.Pod
}

func WatchInitContainerStatus2(pod *apiv1.Pod, containerName string, ch chan<- ContainerExitStatus) {
	for i := 0; i < 12; i++ {
		time.Sleep(5 * time.Second)
		currentPod, err := GetPod(pod.Namespace, pod.Name)
		if err != nil {
			msg := fmt.Sprintf("Cannot get pod %s from namespace %s", pod.Name, pod.Namespace)
			ch <- ContainerExitStatus{
				Success: false,
				Message: msg,
			}
			return
		}
		for _, status := range currentPod.Status.InitContainerStatuses {
			if status.Name != containerName {
				continue
			}
			if status.LastTerminationState.Terminated != nil {
				if status.LastTerminationState.Terminated.ExitCode != 0 {
					msg := fmt.Sprintf("plugin %s failed in pod %s of namespace %s. %s. %s",
						status.Name, pod.Name, pod.Namespace, openapi.ParseErrorMessage(status.LastTerminationState.Terminated.Message).ErrorCode,
						openapi.ParseErrorMessage(status.LastTerminationState.Terminated.Message).Message)
					//log.Error(msg)
					//SendPodEvent(currentPod, apiv1.EventTypeWarning, "Created", msg)
					ch <- ContainerExitStatus{
						Success: false,
						Message: msg,
						Pod:     currentPod,
					}
					return
				}
			}

			if status.State.Terminated != nil {
				if status.State.Terminated.ExitCode == 0 {
					msg := fmt.Sprintf("plugin %s executed successfully in pod %s of namespace %s",
						status.Name, pod.Name, pod.Namespace)
					ch <- ContainerExitStatus{
						Success: true,
						Message: msg,
						Pod:     currentPod,
					}
					return
				}
			}
		}
	}
	msg := fmt.Sprintf("plugin %s takes too long to finish in pod %s of namespace %s",
		containerName, pod.Name, pod.Namespace)
	ch <- ContainerExitStatus{
		Success: false,
		Message: msg,
	}
}

//跟踪pod的init container的启动情况
//func WatchInitContainerStatus(pod *apiv1.Pod, containerName string) {
//	for i := 0; i < 12; i++ {
//		time.Sleep(5 * time.Second)
//		currentPod, err := GetPod(pod.Namespace, pod.Name)
//		if err != nil {
//			log.Error(fmt.Sprintf("Cannot get pod %s from namespace %s", pod.Name, pod.Namespace))
//			return
//		}
//		for _, status := range currentPod.Status.InitContainerStatuses {
//			if status.Name != containerName {
//				continue
//			}
//			if status.LastTerminationState.Terminated != nil {
//				if status.LastTerminationState.Terminated.ExitCode != 0 {
//					msg := fmt.Sprintf("plugin %s failed in pod %s of namespace %s. %s",
//						status.Name, pod.Name, pod.Namespace, openapi.ParseErrorMessage(status.LastTerminationState.Terminated.Message).Message)
//					log.Error(msg)
//					SendPodEvent(currentPod, apiv1.EventTypeWarning, "Created", msg)
//					return
//				}
//			}
//
//			if status.State.Terminated != nil {
//				if status.State.Terminated.ExitCode == 0 {
//					log.Infof("plugin %s executed successfully in pod %s of namespace %s",
//						status.Name, pod.Name, pod.Namespace)
//					return
//				}
//			}
//		}
//	}
//}

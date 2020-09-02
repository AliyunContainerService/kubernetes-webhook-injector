package utils

import (
	"fmt"
	"net"
)

//Watch 给定命名空间下，相同generatedName的，具有pluginName名字的init container的Pod的创建操作，用通道返回
//func WatchPodsWithPlugin(namespace, generateName, pluginName string) <-chan *apiv1.Pod {
//	if podWatcher == nil {
//		initializePodWatcher()
//	}
//	return podWatcher.WatchPlugin(namespace, generateName, pluginName)
//}

//func GetPod(namespace, name string) (*apiv1.Pod, error) {
//	if podWatcher == nil {
//		initializePodWatcher()
//	}
//	return podWatcher.ClientSet.CoreV1().Pods(namespace).Get(context.Background(), name, metav1.GetOptions{})
//}

//func SendEvent(ref runtime.Object, eventType, reason, message string) {
//	if eventor == nil {
//		initializeEventor()
//	}
//	eventor.Recorder.Event(ref, eventType, reason, message)
//}
//
//func SendPodEvent(pod *apiv1.Pod, eventType, reason, message string) {
//	ref, err := reference.GetReference(scheme.Scheme, pod)
//	if err != nil {
//		log.Errorf("Failed to get object reference of pod %s: %v", pod.Name, err)
//	}
//	SendEvent(ref, eventType, reason, message)
//}

func ExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", fmt.Errorf("failed to find external IP address")
}

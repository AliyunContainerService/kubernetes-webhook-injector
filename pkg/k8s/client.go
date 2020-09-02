package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"log"
)

var (
	clientSet kubernetes.Interface
	eventer   *Eventor
)

func GetClientSetOrDie(masterUrl, kubeConfigPath string) kubernetes.Interface {
	config, err := clientcmd.BuildConfigFromFlags(masterUrl, kubeConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}
	return cs
}
func GetClientSet() kubernetes.Interface {
	if clientSet == nil {
		clientSet = GetClientSetOrDie("", "")
	}
	return clientSet
}

func GetEventor() *Eventor {
	if eventer == nil {
		eventer = NewEventor(GetClientSet())
	}
	return eventer
}

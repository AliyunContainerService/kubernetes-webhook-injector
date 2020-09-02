package k8s

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/tools/reference"
	log "k8s.io/klog"
)

type Eventor struct {
	Client   kubernetes.Interface
	Recorder record.EventRecorder
}

func NewEventor(client kubernetes.Interface) *Eventor {
	recorder := newEventRecorder(client)

	return &Eventor{
		Client:   client,
		Recorder: recorder,
	}
}

func newEventRecorder(clientSet kubernetes.Interface) record.EventRecorder {
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(log.Infof)
	eventBroadcaster.StartRecordingToSink(
		&typedcorev1.EventSinkImpl{
			Interface: clientSet.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(
		scheme.Scheme,
		apiv1.EventSource{Component: "controlplane"})
	return recorder
}

func (e *Eventor) SendPodEvent(pod *apiv1.Pod, eventType, reason, message string) {
	ref, err := reference.GetReference(scheme.Scheme, pod)
	if err != nil {
		log.Errorf("Failed to get object reference of pod %s: %v", pod.Name, err)
	}
	e.Recorder.Event(ref, eventType, reason, message)
}

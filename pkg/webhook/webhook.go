package webhook

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/pkg/k8s"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/plugins"
	"io/ioutil"
	"k8s.io/api/admission/v1beta1"
	mutateV1beta1 "k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	log "k8s.io/klog"
	"net/http"
	"strconv"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	// (https://github.com/kubernetes/kubernetes/issues/57982)
	defaulter = runtime.ObjectDefaulter(runtimeScheme)
)

var (
	MutatingWebhookConfigurationName = "kubernetes-webhook-injector"
	MutatingWebhookConfigurationPath = "/mutate"
)

func init() {
	_ = mutateV1beta1.AddToScheme(runtimeScheme)
	// defaulting with webhooks:
	// https://github.com/kubernetes/kubernetes/issues/57982
	_ = v1.AddToScheme(runtimeScheme)
}

// WebHook Server to handle patch request
type WebHookServer struct {
	pluginManager *plugins.PluginManager
	clientSet     kubernetes.Interface
	Options       *WebHookOptions
	Server        *http.Server
}

// Http handler of patch request
func (ws *WebHookServer) Serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		log.Error("Empty body of patch body.")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		log.Errorf("Content-Type=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	// decode response
	var admissionResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		log.Errorf("Can't decode body: %v", err)
		admissionResponse = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		// handle path and return mutate response
		if r.URL.Path == "/mutate" {
			admissionResponse = ws.mutate(&ar)
		}
	}

	// wrapper admissionReview response
	admissionReview := v1beta1.AdmissionReview{}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		log.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}

	if _, err := w.Write(resp); err != nil {
		log.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}

// mutate the pod spec and patch pod
func (ws *WebHookServer) mutate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	// default log level is 2
	log.V(5).Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v (%v) UID=%v patchOperation=%v UserInfo=%v",
		req.Kind, req.Namespace, req.Name, req.Object, req.UID, req.Operation, req.UserInfo)

	raw := req.Object.Raw
	pod := &v1.Pod{}

	// Create ??????????????? req.Object.Raw, ???????????????????????????
	if req.Operation == v1beta1.Create {
		if err := json.Unmarshal(raw, pod); err != nil {
			log.Errorf("Failed to unmarshal pod %v,because of %v", raw, err)
			return &v1beta1.AdmissionResponse{
				Allowed: true,
			}
		}
	}

	// ????????? DELETE ??????????????????, ????????????????????????????????????operation?????????????????????????????????????????????????????????,?????????????????????plugin manager???
	pod.Namespace = req.Namespace
	pod.Name = req.Name
	if req.Operation == v1beta1.Delete {
		p, err := ws.clientSet.CoreV1().Pods(pod.Namespace).Get(context.Background(), pod.Name, metav1.GetOptions{})
		if err != nil {
			log.Errorf("failed to get pod %s from namespace %s ,because of %v", pod.Name, pod.Namespace, err)
			return &v1beta1.AdmissionResponse{
				Allowed: true,
			}
		}
		pod = p
	}

	patchBytes, err := ws.pluginManager.HandlePatchPod(pod, req.Operation)
	if err != nil {
		log.Errorf("Failed to patch pod %v,because of %v", pod, err)
		return &v1beta1.AdmissionResponse{
			Allowed: true,
		}
	}

	if patchBytes != nil {
		response := &v1beta1.AdmissionResponse{Allowed: true}
		response.Patch = patchBytes
		patchType := v1beta1.PatchTypeJSONPatch
		response.PatchType = &patchType
		// change patch debug log level to 5
		log.V(5).Infof("Successfully patch pod %s in %s with pathOps %v", pod.Name, pod.Namespace, string(patchBytes))
		return response
	}

	return &v1beta1.AdmissionResponse{
		Allowed: true,
	}
}

// register MutatingWebHookConfiguration
func (ws *WebHookServer) registerMutatingWebhookConfiguration() error {
	mutatingConfigs := ws.clientSet.AdmissionregistrationV1beta1().MutatingWebhookConfigurations()
	conf, err := mutatingConfigs.Get(context.Background(), MutatingWebhookConfigurationName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// todo create a new one
			mutatingRules := []mutateV1beta1.RuleWithOperations{
				{
					Operations: []mutateV1beta1.OperationType{mutateV1beta1.Create, mutateV1beta1.Delete},
					Rule: mutateV1beta1.Rule{
						APIGroups:   []string{""},
						APIVersions: []string{"v1"},
						Resources:   []string{"pods"},
					},
				},
			}

			// read ca cert data from path
			caCert, err := ioutil.ReadFile(ws.Options.TLSCaCertPath)
			if err != nil {
				return err
			}

			// parse service port to int32 pointer
			port, err := strconv.ParseInt(ws.Options.Port, 10, 32)
			if err != nil {
				return err
			}
			portInt32 := int32(port)

			mutatingWebHook := mutateV1beta1.MutatingWebhook{
				Name:  "kubernetes-webhook-injector.ack.aliyun.com",
				Rules: mutatingRules,
				ClientConfig: mutateV1beta1.WebhookClientConfig{
					Service: &mutateV1beta1.ServiceReference{
						Namespace: ws.Options.ServiceNamespace,
						Name:      ws.Options.ServiceName,
						Port:      &portInt32,
						Path:      &MutatingWebhookConfigurationPath,
					},
					CABundle: caCert,
				},
			}

			webhookConfig := &mutateV1beta1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: MutatingWebhookConfigurationName,
				},
				Webhooks: []mutateV1beta1.MutatingWebhook{mutatingWebHook},
			}

			if _, err := mutatingConfigs.Create(context.Background(), webhookConfig, metav1.CreateOptions{}); err != nil {
				log.Errorf("Failed to create MutatingWebhookConfiguration %s,because of %v", MutatingWebhookConfigurationName, err)
				return err
			}
		}
		log.Errorf("Failed to get MutatingWebhookConfiguration %s,because of %v", MutatingWebhookConfigurationName, err)
		return err
	}
	if conf != nil {
		log.Infof("MutatingWebhookConfiguration %s has been created", MutatingWebhookConfigurationName)
	}
	return nil
}

// register MutatingWebhookConfiguration and serve the request
func (ws *WebHookServer) Run() (err error) {
	if err = ws.registerMutatingWebhookConfiguration(); err != nil {
		log.Errorf("Failed to register MutatingWebhookConfiguration,because of %v", err)
		return err
	}
	return ws.Server.ListenAndServeTLS("", "")
}

// NewWebHookServer return mutate web server
func NewWebHookServer(wo *WebHookOptions) (ws *WebHookServer, err error) {

	//if wo.KubeConf == "" {
	//	if home := homedir.HomeDir(); home != "" {
	//		wo.KubeConf = filepath.Join(home, ".kube", "config")
	//	}
	//}

	// ????????????????????????inClusterConfig:
	//tokenFile  = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	//rootCAFile = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	k8s.InitClientSetOrDie("", wo.KubeConf)

	ws = &WebHookServer{
		clientSet:     k8s.GetClientSet(),
		Options:       wo,
		pluginManager: plugins.NewPluginManager(),
		Server: &http.Server{
			Addr:      fmt.Sprintf(":%v", wo.Port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{wo.TLSPair}},
		},
	}
	return ws, nil
}

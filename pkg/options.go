package pkg

import (
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	log "k8s.io/klog"
	"os"
)

type WebHookOptions struct {
	// TLS key and value
	TLSCertPath   string
	TLSKeyPath    string
	TLSCaCertPath string

	TLSPair tls.Certificate
	// Server Port
	Port string
	//service configuration
	ServiceName      string
	ServiceNamespace string
	// leader election option
	LeaderElection bool
	// kubeconf path
	KubeConf string
}

// NewWebHookOptions parse the command line params and initialize the server
func NewWebHookOptions() (options *WebHookOptions, err error) {
	wo := &WebHookOptions{}
	// initialize the flag parse
	wo.init()

	// todo add strict validation [Empty/Pattern]
	if passed, msg := wo.valid(); !passed {
		log.Errorf("Failed to pass webHook options validation,because of %v", msg)
		return nil, errors.New(msg)
	}

	return wo, nil
}

// init flag params and parse
func (wo *WebHookOptions) init() {
	// tls configurations
	flag.StringVar(&wo.TLSCaCertPath, "ca", "/run/secrets/tls/ca-cert.pem", "The path of ca cert.")
	flag.StringVar(&wo.TLSCertPath, "cert", "/run/secrets/tls/server-cert.pem", "The path of TLS cert.")
	flag.StringVar(&wo.TLSKeyPath, "key", "/run/secrets/tls/server-key.pem", "The path of TLS key.")

	flag.StringVar(&wo.ServiceName, "service-name", "kubernetes-webhook-injector", "The service of kubernetes-webhook-injector.")
	flag.StringVar(&wo.ServiceNamespace, "service-namespace", "kube-system", "The namespace of kubernetes-webhook-injector.")
	flag.StringVar(&wo.Port, "port", "443", "The webhook service port of kubernetes-webhook-injector.")

	flag.StringVar(&wo.KubeConf, "kubeconf", "", "use ~/.kube/conf as default.")
	// todo enable leader election to support high performance
	flag.BoolVar(&wo.LeaderElection, "leaderElection", true, "Enable leaderElection or not.")

	flag.Parse()
}

// check params is valid or not
func (wo *WebHookOptions) valid() (passed bool, msg string) {

	// check file exist or not
	if _, err := os.Stat(wo.TLSCertPath); err != nil && os.IsNotExist(err) {
		return false, fmt.Sprintf("TLSCert is not found.")
	}

	// load key pair from file
	pair, err := tls.LoadX509KeyPair(wo.TLSCertPath, wo.TLSKeyPath)
	if err != nil {
		return false, fmt.Sprintf("Failed to parse certificate,because of %v", err)
	}
	wo.TLSPair = pair

	// todo add other validations
	// code block

	return true, ""
}

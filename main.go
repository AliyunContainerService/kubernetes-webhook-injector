/*
Copyright (c) 2019 StackRox Inc.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
   http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"github.com/AliyunContainerService/kubernetes-webhook-injector/pkg/openapi"
	"github.com/AliyunContainerService/kubernetes-webhook-injector/pkg/webhook"
	"log"
	"net/http"
)

func main() {
	var wo *webhook.WebHookOptions
	var err error
	if wo, err = webhook.NewWebHookOptions(); err != nil {
		log.Fatalf("Please input valid params. %v", err)
	}
	openapi.InitClient(wo.IntranetAccess)

	ws, err := webhook.NewWebHookServer(wo)

	if err != nil {
		log.Fatalf("Failed to set up webhook server: %v", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc(webhook.MutatingWebhookConfigurationPath, ws.Serve)
	ws.Server.Handler = mux

	log.Fatal(ws.Run())
}

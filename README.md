## Introduction
In cloud scenarios with  relatively high permission requirements, the IP address of Pod needs to be dynamically added to or removed from the specified Alibaba cloud product whitelist to achieve the most fine-grained control over permissions. You can use ack-kubernetes-webhook-injector to dynamically add or remove Pod IP from the whitelist by annotating the Pod.

ack-kubernetes-webhook-injector is a kubernetes component that can dynamically add or remove Pod IP from multiple Alibaba cloud product whitelists, eliminating the need to manually configure Pod IP to the cloud product whitelist. 

Currently ack-kubernetes-webhook-injector supports the following functions:

* When creating/deleting a Pod, automatically add/remove the Pod's IP from the specified RDS whitelist .
* When creating/deleting a Pod, automatically add/remove the Pod's IP from the specified REDIS whitelist.
* When creating/deleting a Pod, automatically add/remove the Pod's IP from the specified SLB access control list.

## Quick Start
### Build and Deploy
1. Build the project
```
# make
```
2. Deploy the project
```
# make docker-build
```
### Uninstall
If you no longer need the ack-kubernetes-webhook-injector, please execute the following command to clear the configuration information while deleting it:
```
# kubectl -n kube-system delete secret kubernetes-webhook-injector-certs
# kubectl delete mutatingwebhookconfigurations.admissionregistration.k8s.io kubernetes-webhook-injector
```
## Configuration and Use
You only need to use Annotation in the Pod Spec of the Pod replica controller to indicate the RDS instance ID and the RDS whitelist grouping name. When a Pod is created, ack-kubernetes-webhook-injector adds the IP address of the Pod to a whitelist or security group rule and removes the rule when the Pod is deleted.
Pod's Annotation needs to include:
* RDS Whitelist:
  - RDS Instance ID: `ack.aliyun.com/rds_id`
  - RDS Whitelist Group Name: `ack.aliyun.com/white_list_name`
* SLB Access Control: `ack.aliyun.com/access_control_policy_id`
* Redis Whitelist:
  - Redis Instance ID: `ack.aliyun.com/redis_id`
  - Redis White List Grouping: `ack.aliyun.com/redis_white_list_name`

For more information, see [Configure the Alibaba cloud product whitelist dynamically for Pod](https://help.aliyun.com/document_detail/188574.html).
## License
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at
``
http://www.apache.org/licenses/LICENSE-2.0
``
.

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

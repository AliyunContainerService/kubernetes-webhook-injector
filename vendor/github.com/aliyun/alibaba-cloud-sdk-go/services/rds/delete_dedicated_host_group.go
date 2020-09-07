package rds

//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//
// Code generated by Alibaba Cloud SDK Code Generator.
// Changes may cause incorrect behavior and will be lost if the code is regenerated.

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/responses"
)

// DeleteDedicatedHostGroup invokes the rds.DeleteDedicatedHostGroup API synchronously
// api document: https://help.aliyun.com/api/rds/deletededicatedhostgroup.html
func (client *Client) DeleteDedicatedHostGroup(request *DeleteDedicatedHostGroupRequest) (response *DeleteDedicatedHostGroupResponse, err error) {
	response = CreateDeleteDedicatedHostGroupResponse()
	err = client.DoAction(request, response)
	return
}

// DeleteDedicatedHostGroupWithChan invokes the rds.DeleteDedicatedHostGroup API asynchronously
// api document: https://help.aliyun.com/api/rds/deletededicatedhostgroup.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DeleteDedicatedHostGroupWithChan(request *DeleteDedicatedHostGroupRequest) (<-chan *DeleteDedicatedHostGroupResponse, <-chan error) {
	responseChan := make(chan *DeleteDedicatedHostGroupResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DeleteDedicatedHostGroup(request)
		if err != nil {
			errChan <- err
		} else {
			responseChan <- response
		}
	})
	if err != nil {
		errChan <- err
		close(responseChan)
		close(errChan)
	}
	return responseChan, errChan
}

// DeleteDedicatedHostGroupWithCallback invokes the rds.DeleteDedicatedHostGroup API asynchronously
// api document: https://help.aliyun.com/api/rds/deletededicatedhostgroup.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DeleteDedicatedHostGroupWithCallback(request *DeleteDedicatedHostGroupRequest, callback func(response *DeleteDedicatedHostGroupResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DeleteDedicatedHostGroupResponse
		var err error
		defer close(result)
		response, err = client.DeleteDedicatedHostGroup(request)
		callback(response, err)
		result <- 1
	})
	if err != nil {
		defer close(result)
		callback(nil, err)
		result <- 0
	}
	return result
}

// DeleteDedicatedHostGroupRequest is the request struct for api DeleteDedicatedHostGroup
type DeleteDedicatedHostGroupRequest struct {
	*requests.RpcRequest
	ResourceOwnerId      requests.Integer `position:"Query" name:"ResourceOwnerId"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
	DedicatedHostGroupId string           `position:"Query" name:"DedicatedHostGroupId"`
}

// DeleteDedicatedHostGroupResponse is the response struct for api DeleteDedicatedHostGroup
type DeleteDedicatedHostGroupResponse struct {
	*responses.BaseResponse
	RequestId string `json:"RequestId" xml:"RequestId"`
}

// CreateDeleteDedicatedHostGroupRequest creates a request to invoke DeleteDedicatedHostGroup API
func CreateDeleteDedicatedHostGroupRequest() (request *DeleteDedicatedHostGroupRequest) {
	request = &DeleteDedicatedHostGroupRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Rds", "2014-08-15", "DeleteDedicatedHostGroup", "rds", "openAPI")
	request.Method = requests.POST
	return
}

// CreateDeleteDedicatedHostGroupResponse creates a response to parse from DeleteDedicatedHostGroup response
func CreateDeleteDedicatedHostGroupResponse() (response *DeleteDedicatedHostGroupResponse) {
	response = &DeleteDedicatedHostGroupResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}
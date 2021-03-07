package r_kvstore

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

// DescribeLogicInstanceTopology invokes the r_kvstore.DescribeLogicInstanceTopology API synchronously
// api document: https://help.aliyun.com/api/r-kvstore/describelogicinstancetopology.html
func (client *Client) DescribeLogicInstanceTopology(request *DescribeLogicInstanceTopologyRequest) (response *DescribeLogicInstanceTopologyResponse, err error) {
	response = CreateDescribeLogicInstanceTopologyResponse()
	err = client.DoAction(request, response)
	return
}

// DescribeLogicInstanceTopologyWithChan invokes the r_kvstore.DescribeLogicInstanceTopology API asynchronously
// api document: https://help.aliyun.com/api/r-kvstore/describelogicinstancetopology.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeLogicInstanceTopologyWithChan(request *DescribeLogicInstanceTopologyRequest) (<-chan *DescribeLogicInstanceTopologyResponse, <-chan error) {
	responseChan := make(chan *DescribeLogicInstanceTopologyResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DescribeLogicInstanceTopology(request)
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

// DescribeLogicInstanceTopologyWithCallback invokes the r_kvstore.DescribeLogicInstanceTopology API asynchronously
// api document: https://help.aliyun.com/api/r-kvstore/describelogicinstancetopology.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeLogicInstanceTopologyWithCallback(request *DescribeLogicInstanceTopologyRequest, callback func(response *DescribeLogicInstanceTopologyResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DescribeLogicInstanceTopologyResponse
		var err error
		defer close(result)
		response, err = client.DescribeLogicInstanceTopology(request)
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

// DescribeLogicInstanceTopologyRequest is the request struct for api DescribeLogicInstanceTopology
type DescribeLogicInstanceTopologyRequest struct {
	*requests.RpcRequest
	ResourceOwnerId      requests.Integer `position:"Query" name:"ResourceOwnerId"`
	SecurityToken        string           `position:"Query" name:"SecurityToken"`
	ResourceOwnerAccount string           `position:"Query" name:"ResourceOwnerAccount"`
	OwnerAccount         string           `position:"Query" name:"OwnerAccount"`
	OwnerId              requests.Integer `position:"Query" name:"OwnerId"`
	InstanceId           string           `position:"Query" name:"InstanceId"`
}

// DescribeLogicInstanceTopologyResponse is the response struct for api DescribeLogicInstanceTopology
type DescribeLogicInstanceTopologyResponse struct {
	*responses.BaseResponse
	RequestId      string         `json:"RequestId" xml:"RequestId"`
	InstanceId     string         `json:"InstanceId" xml:"InstanceId"`
	RedisProxyList RedisProxyList `json:"RedisProxyList" xml:"RedisProxyList"`
	RedisShardList RedisShardList `json:"RedisShardList" xml:"RedisShardList"`
}

// CreateDescribeLogicInstanceTopologyRequest creates a request to invoke DescribeLogicInstanceTopology API
func CreateDescribeLogicInstanceTopologyRequest() (request *DescribeLogicInstanceTopologyRequest) {
	request = &DescribeLogicInstanceTopologyRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("R-kvstore", "2015-01-01", "DescribeLogicInstanceTopology", "redisa", "openAPI")
	request.Method = requests.POST
	return
}

// CreateDescribeLogicInstanceTopologyResponse creates a response to parse from DescribeLogicInstanceTopology response
func CreateDescribeLogicInstanceTopologyResponse() (response *DescribeLogicInstanceTopologyResponse) {
	response = &DescribeLogicInstanceTopologyResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}

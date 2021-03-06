package drds

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

// DescribeCanExpandInstanceDetailList invokes the drds.DescribeCanExpandInstanceDetailList API synchronously
// api document: https://help.aliyun.com/api/drds/describecanexpandinstancedetaillist.html
func (client *Client) DescribeCanExpandInstanceDetailList(request *DescribeCanExpandInstanceDetailListRequest) (response *DescribeCanExpandInstanceDetailListResponse, err error) {
	response = CreateDescribeCanExpandInstanceDetailListResponse()
	err = client.DoAction(request, response)
	return
}

// DescribeCanExpandInstanceDetailListWithChan invokes the drds.DescribeCanExpandInstanceDetailList API asynchronously
// api document: https://help.aliyun.com/api/drds/describecanexpandinstancedetaillist.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeCanExpandInstanceDetailListWithChan(request *DescribeCanExpandInstanceDetailListRequest) (<-chan *DescribeCanExpandInstanceDetailListResponse, <-chan error) {
	responseChan := make(chan *DescribeCanExpandInstanceDetailListResponse, 1)
	errChan := make(chan error, 1)
	err := client.AddAsyncTask(func() {
		defer close(responseChan)
		defer close(errChan)
		response, err := client.DescribeCanExpandInstanceDetailList(request)
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

// DescribeCanExpandInstanceDetailListWithCallback invokes the drds.DescribeCanExpandInstanceDetailList API asynchronously
// api document: https://help.aliyun.com/api/drds/describecanexpandinstancedetaillist.html
// asynchronous document: https://help.aliyun.com/document_detail/66220.html
func (client *Client) DescribeCanExpandInstanceDetailListWithCallback(request *DescribeCanExpandInstanceDetailListRequest, callback func(response *DescribeCanExpandInstanceDetailListResponse, err error)) <-chan int {
	result := make(chan int, 1)
	err := client.AddAsyncTask(func() {
		var response *DescribeCanExpandInstanceDetailListResponse
		var err error
		defer close(result)
		response, err = client.DescribeCanExpandInstanceDetailList(request)
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

// DescribeCanExpandInstanceDetailListRequest is the request struct for api DescribeCanExpandInstanceDetailList
type DescribeCanExpandInstanceDetailListRequest struct {
	*requests.RpcRequest
	CurrentPlan    string `position:"Query" name:"CurrentPlan"`
	DbName         string `position:"Query" name:"DbName"`
	DrdsInstanceId string `position:"Query" name:"DrdsInstanceId"`
}

// DescribeCanExpandInstanceDetailListResponse is the response struct for api DescribeCanExpandInstanceDetailList
type DescribeCanExpandInstanceDetailListResponse struct {
	*responses.BaseResponse
	RequestId string                                    `json:"RequestId" xml:"RequestId"`
	Success   bool                                      `json:"Success" xml:"Success"`
	Data      DataInDescribeCanExpandInstanceDetailList `json:"Data" xml:"Data"`
}

// CreateDescribeCanExpandInstanceDetailListRequest creates a request to invoke DescribeCanExpandInstanceDetailList API
func CreateDescribeCanExpandInstanceDetailListRequest() (request *DescribeCanExpandInstanceDetailListRequest) {
	request = &DescribeCanExpandInstanceDetailListRequest{
		RpcRequest: &requests.RpcRequest{},
	}
	request.InitWithApiInfo("Drds", "2019-01-23", "DescribeCanExpandInstanceDetailList", "drds", "openAPI")
	return
}

// CreateDescribeCanExpandInstanceDetailListResponse creates a response to parse from DescribeCanExpandInstanceDetailList response
func CreateDescribeCanExpandInstanceDetailListResponse() (response *DescribeCanExpandInstanceDetailListResponse) {
	response = &DescribeCanExpandInstanceDetailListResponse{
		BaseResponse: &responses.BaseResponse{},
	}
	return
}

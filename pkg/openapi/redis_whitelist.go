package openapi

import (
	"fmt"
	redis "github.com/aliyun/alibaba-cloud-sdk-go/services/r-kvstore"
)

func (ro *RedisWhitelistOperator) DescribeWhitelists(redisId string) ([]redis.SecurityIpGroup, error) {

	request := redis.CreateDescribeSecurityIpsRequest()
	request.InstanceId = redisId
	request.Scheme = "https"
	request.RegionId = RegionID

	response, err := ro.DescribeSecurityIps(request)
	if err != nil {
		return nil, err
	}

	return response.SecurityIpGroups.SecurityIpGroup, nil
}

func (ro *RedisWhitelistOperator) CreateOrAppendWhitelist(redisId, podIP, whitelistName string) error {

	request := redis.CreateModifySecurityIpsRequest()
	request.Scheme = "https"
	request.SecurityIps = podIP
	request.InstanceId = redisId
	request.ModifyMode = "Append"

	if whitelistName == "" {
		request.SecurityIpGroupName = "default"
	} else {
		request.SecurityIpGroupName = whitelistName
	}

	_, err := ro.ModifySecurityIps(request)
	return err
}

func (ro *RedisWhitelistOperator) GetWhiteList(redisId, whitelistName string) (*redis.SecurityIpGroup, error) {
	whitelists, err := ro.DescribeWhitelists(redisId)
	if err != nil {
		return nil, err
	}

	for _, wl := range whitelists {
		if wl.SecurityIpGroupName == whitelistName {
			return &wl, nil
		}
	}

	return nil, nil
}

func (ro *RedisWhitelistOperator) DeleteWhitelist(redisId, whitelistName, podIP string) error {

	list, err := ro.GetWhiteList(redisId, whitelistName)
	if err != nil {
		return err
	}

	if list == nil {
		return fmt.Errorf("no such whitelist %s under redisId %s", whitelistName, redisId)
	}

	request := redis.CreateModifySecurityIpsRequest()
	request.Scheme = "https"
	request.RegionId = RegionID
	request.InstanceId = redisId
	request.SecurityIps = podIP
	request.ModifyMode = "Delete"

	if whitelistName == "" {
		request.SecurityIpGroupName = "default"
	} else {
		request.SecurityIpGroupName = whitelistName
	}

	_, err = ro.ModifySecurityIps(request)
	return err
}

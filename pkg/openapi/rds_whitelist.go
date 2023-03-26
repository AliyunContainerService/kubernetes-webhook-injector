package openapi

import (
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
)

func (r *RdsWhitelistOperator) Whitelists(rdsId string) ([]rds.DBInstanceIPArray, error) {

	var lists []rds.DBInstanceIPArray
	req := rds.CreateDescribeDBInstanceIPArrayListRequest()
	req.RegionId = RegionID
	req.DBInstanceId = rdsId

	networkTypes := []string{"MIX", "Classic", "VPC"}
	for _, networkType := range networkTypes {
		req.WhitelistNetworkType = networkType
		resp, err := r.DescribeDBInstanceIPArrayList(req)
		if err != nil {
			return nil, err
		}

		for _, ipArray := range resp.Items.DBInstanceIPArray {
			lists = append(lists, ipArray)
		}
	}

	return lists, nil
}

func (r *RdsWhitelistOperator) CreateWhitelist(rdsId, podIP, whitelistName string) error {

	lists, err := r.GetWhitelist(rdsId, whitelistName)
	if err != nil {
		return err
	}

	if lists == nil {
		return fmt.Errorf("no such whitelist %s under rdsId %s", whitelistName, rdsId)
	}

	for _, list := range lists {
		req := rds.CreateModifySecurityIpsRequest()
		req.RegionId = RegionID
		req.DBInstanceId = rdsId
		req.SecurityIPType = "ipv4"
		req.SecurityIps = podIP
		req.WhitelistNetworkType = list.WhitelistNetworkType
		req.ModifyMode = "Append"

		if whitelistName == "" {
			req.DBInstanceIPArrayName = "rds_default"
		} else {
			req.DBInstanceIPArrayName = whitelistName
		}

		_, err := r.ModifySecurityIps(req)

		if err != nil {
			if ParseErrorMessage(err.Error()).ErrorCode == "InvalidInstanceIp.Duplicate" {
				continue
			}
			return err
		}
	}
	return nil
}

func (r *RdsWhitelistOperator) GetWhitelist(rdsId, whitelistName string) ([]rds.DBInstanceIPArray, error) {

	lists, err := r.Whitelists(rdsId)
	if err != nil {
		return nil, err
	}

	var whitelists []rds.DBInstanceIPArray
	for _, list := range lists {
		if list.DBInstanceIPArrayName == whitelistName {
			whitelists = append(whitelists, list)
		}
	}

	return whitelists, nil
}

func (r *RdsWhitelistOperator) DeleteWhitelist(rdsId, whitelistName, podIP string) error {

	lists, err := r.GetWhitelist(rdsId, whitelistName)
	if err != nil {
		return err
	}

	if lists == nil {
		return fmt.Errorf("no such whitelist %s under rdsId %s", whitelistName, rdsId)
	}

	for _, list := range lists {
		req := rds.CreateModifySecurityIpsRequest()
		req.RegionId = RegionID
		req.DBInstanceId = rdsId
		req.WhitelistNetworkType = list.WhitelistNetworkType
		req.SecurityIps = podIP
		req.ModifyMode = "Delete"

		if whitelistName == "" {
			req.DBInstanceIPArrayName = "default"
		} else {
			req.DBInstanceIPArrayName = whitelistName
		}

		_, err = r.ModifySecurityIps(req)
		if err != nil {
			if ParseErrorMessage(err.Error()).ErrorCode == "InvalidSecurityIPs.NotFound" {
				continue
			}
			return err
		}
	}

	return nil
}

//func RefactorRdsWhitelistName(name string) string {
//	//rds对白名单名字的要求：由小写字母、数字、下划线组成，以小写字母开头，以字母或数字结尾。长度为2-32个字符
//	newName := strings.ReplaceAll(name, "-", "_")
//	if len(newName) > 32 {
//		newName = newName[len(newName)-32:]
//	}
//	if strings.HasPrefix(newName, "_") {
//		newName = newName[1:]
//	}
//
//	newName = "a" + newName[1:]
//	return newName
//}

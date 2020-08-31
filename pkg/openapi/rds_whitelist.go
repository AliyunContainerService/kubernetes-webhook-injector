package openapi

import (
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
	"strings"
)

func (r *RdsWhitelistOperator) Whitelists(rdsId string) ([]rds.DBInstanceIPArray, error) {

	req := rds.CreateDescribeDBInstanceIPArrayListRequest()
	req.RegionId = RegionID
	req.DBInstanceId = rdsId

	resp, err := r.DescribeDBInstanceIPArrayList(req)
	if err != nil {
		return nil, err
	}
	return resp.Items.DBInstanceIPArray, nil
}

func (r *RdsWhitelistOperator) CreateWhitelist(rdsId, podIP, whitelistName string) error {

	req := rds.CreateModifySecurityIpsRequest()
	req.RegionId = RegionID
	req.DBInstanceId = rdsId
	req.SecurityIPType = "ipv4"
	req.WhitelistNetworkType = "vpc"
	req.DBInstanceIPArrayName = whitelistName
	req.SecurityIps = podIP

	_, err := r.ModifySecurityIps(req)

	return err
}

func (r *RdsWhitelistOperator) GetWhitelist(rdsId, whitelistName string) (*rds.DBInstanceIPArray, error) {

	lists, err := r.Whitelists(rdsId)
	if err != nil {
		return nil, err
	}

	for _, list := range lists {
		if list.DBInstanceIPArrayName == whitelistName {
			return &list, nil
		}
	}

	return nil, nil
}

func (r *RdsWhitelistOperator) DeleteWhitelist(rdsId, whitelistName string) error {

	list, err := r.GetWhitelist(rdsId, whitelistName)
	if err != nil {
		return err
	}

	if list == nil {
		return fmt.Errorf("no such whitelist %s under rdsId %s", whitelistName, rdsId)
	}

	req := rds.CreateModifySecurityIpsRequest()
	req.RegionId = RegionID
	req.DBInstanceId = rdsId
	req.DBInstanceIPArrayName = whitelistName
	req.SecurityIps = list.SecurityIPList
	req.WhitelistNetworkType = list.WhitelistNetworkType
	req.ModifyMode = "Delete"

	_, err = r.ModifySecurityIps(req)
	return err
}

func RefactorRdsWhitelistName(name string) string {
	//rds对白名单名字的要求：由小写字母、数字、下划线组成，以小写字母开头，以字母或数字结尾。长度为2-32个字符
	newName := strings.ReplaceAll(name, "-", "_")
	if len(newName) > 32 {
		newName = newName[len(newName)-32:]
	}
	if strings.HasPrefix(newName, "_") {
		newName = newName[1:]
	}
	return newName
}

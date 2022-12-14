package openapi

import (
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/rds"
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
	req.SecurityIps = podIP
	req.WhitelistNetworkType = "MIX"
	req.ModifyMode = "Append"

	if whitelistName == "" {
		req.DBInstanceIPArrayName = "rds_default"
	} else {
		req.DBInstanceIPArrayName = whitelistName
	}

	_, err := r.ModifySecurityIps(req)

	if err != nil && ParseErrorMessage(err.Error()).ErrorCode == "SecurityIPList.Duplicate" {
		return nil
	}

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

func (r *RdsWhitelistOperator) DeleteWhitelist(rdsId, whitelistName, podIP string) error {

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
	req.SecurityIps = list.SecurityIPList
	req.WhitelistNetworkType = list.WhitelistNetworkType
	req.SecurityIps = podIP
	req.ModifyMode = "Delete"

	if whitelistName == "" {
		req.DBInstanceIPArrayName = "default"
	} else {
		req.DBInstanceIPArrayName = whitelistName
	}

	_, err = r.ModifySecurityIps(req)
	return err
}

//func RefactorRdsWhitelistName(name string) string {
//	//rds???????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????2-32?????????
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

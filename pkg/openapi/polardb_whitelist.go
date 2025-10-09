package openapi

import (
	"fmt"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/polardb"
)

func (o *PolarDBWhitelistOperator) Whitelists(polardbId string) ([]polardb.DBClusterIPArray, error) {

	var lists []polardb.DBClusterIPArray
	req := polardb.CreateDescribeDBClusterAccessWhitelistRequest()
	req.RegionId = RegionID
	req.DBClusterId = polardbId

	resp, err := o.DescribeDBClusterAccessWhitelist(req)
	if err != nil {
		return nil, err
	}

	lists = append(lists, resp.Items.DBClusterIPArray...)

	return lists, nil
}

func (o *PolarDBWhitelistOperator) CreateWhitelist(polardbId, podIP, whitelistName string) error {

	lists, err := o.GetWhitelist(polardbId, whitelistName)
	if err != nil {
		return err
	}

	if lists == nil {
		return fmt.Errorf("no such whitelist %s under polardbId %s", whitelistName, polardbId)
	}

	for range lists {
		req := polardb.CreateModifyDBClusterAccessWhitelistRequest()
		req.RegionId = RegionID
		req.DBClusterId = polardbId
		req.SecurityIps = podIP
		req.ModifyMode = "Append"

		if whitelistName == "" {
			req.DBClusterIPArrayName = "default"
		} else {
			req.DBClusterIPArrayName = whitelistName
		}

		_, err := o.ModifyDBClusterAccessWhitelist(req)

		if err != nil {
			if ParseErrorMessage(err.Error()).ErrorCode == "ErrorCode: InvalidInstanceIp.Duplicate" {
				continue
			}
			return err
		}
	}
	return nil
}

func (o *PolarDBWhitelistOperator) GetWhitelist(polardbId, whitelistName string) ([]polardb.DBClusterIPArray, error) {

	lists, err := o.Whitelists(polardbId)
	if err != nil {
		return nil, err
	}

	var whitelists []polardb.DBClusterIPArray
	for _, list := range lists {
		if list.DBClusterIPArrayName == whitelistName {
			whitelists = append(whitelists, list)
		}
	}

	return whitelists, nil
}

func (o *PolarDBWhitelistOperator) DeleteWhitelist(polardbID, whitelistName, podIP string) error {

	lists, err := o.GetWhitelist(polardbID, whitelistName)
	if err != nil {
		return err
	}

	if lists == nil {
		return fmt.Errorf("no such whitelist %s under polardbID %s", whitelistName, polardbID)
	}

	for range lists {
		req := polardb.CreateModifyDBClusterAccessWhitelistRequest()
		req.RegionId = RegionID
		req.DBClusterId = polardbID
		req.SecurityIps = podIP
		req.ModifyMode = "Delete"

		if whitelistName == "" {
			req.DBClusterIPArrayName = "default"
		} else {
			req.DBClusterIPArrayName = whitelistName
		}

		_, err = o.ModifyDBClusterAccessWhitelist(req)
		if err != nil {
			if ParseErrorMessage(err.Error()).ErrorCode == "ErrorCode: InvalidSecurityIPs.NotFound" {
				continue
			}
			return err
		}
	}

	return nil
}

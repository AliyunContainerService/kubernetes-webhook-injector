package openapi

import (
	"fmt"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
)

func (so *SLBAccessControlPolicyOperator) AddEntryToAccessControlList(AclId, podIP string) error {

	entry := fmt.Sprintf(`[{"entry":"%s/32"}]`, podIP)

	request := slb.CreateAddAccessControlListEntryRequest()
	request.Scheme = "https"
	request.RegionId = RegionID
	request.AclId = AclId
	request.AclEntrys = entry

	_, err := so.AddAccessControlListEntry(request)
	return err
}

func (so *SLBAccessControlPolicyOperator) DeleteEntryFromAccessControlPolicy(AclId, podIP string) error {

	entry := fmt.Sprintf(`[{"entry":"%s/32"}]`, podIP)

	request := slb.CreateRemoveAccessControlListEntryRequest()
	request.Scheme = "https"
	request.RegionId = RegionID
	request.AclId = AclId
	request.AclEntrys = entry

	_, err := so.RemoveAccessControlListEntry(request)
	return err
}

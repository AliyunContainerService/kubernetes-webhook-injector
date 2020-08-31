package openapi

import "github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"

func (s *SecurityGroupOperator) Permissions(sgID string) ([]ecs.Permission, error) {

	req := ecs.CreateDescribeSecurityGroupAttributeRequest()
	req.RegionId = RegionID
	req.SecurityGroupId = sgID
	resp, err := s.DescribeSecurityGroupAttribute(req)
	if err != nil {
		return nil, err
	}

	return resp.Permissions.Permission, nil
}

func (s *SecurityGroupOperator) CreatePermission(sgID, sourcIP, description string) error {
	req := ecs.CreateAuthorizeSecurityGroupRequest()

	req.RegionId = RegionID
	req.SecurityGroupId = sgID
	req.PortRange = "-1/-1"
	req.IpProtocol = "all"
	req.Description = description
	req.SourceCidrIp = sourcIP + "/32"

	_, err := s.AuthorizeSecurityGroup(req)
	return err
}

func (s *SecurityGroupOperator) FindPermission(sgID, description string) (*ecs.Permission, error) {
	permissions, err := s.Permissions(sgID)
	if err != nil {
		return nil, err
	}

	for _, p := range permissions {
		if p.Description == description {
			return &p, nil
		}
	}

	//没找到不等于有error
	return nil, nil
}

func (s *SecurityGroupOperator) DeletePermission(sgID string, p *ecs.Permission) error {
	req := ecs.CreateRevokeSecurityGroupRequest()
	req.SecurityGroupId = sgID
	req.PortRange = p.PortRange
	req.IpProtocol = p.IpProtocol
	req.SourceCidrIp = p.SourceCidrIp
	req.Description = p.Description

	_, err := s.RevokeSecurityGroup(req)
	return err
}

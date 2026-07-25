package main

import (
	"context"

	basemodel "shop/pkg/model"
	"shop/pkg/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Address 收货地址：按 user_id 归属。
type Address struct {
	basemodel.BaseModel
	UserID       int32  `gorm:"index;not null"`
	Province     string `gorm:"type:varchar(20);not null"`
	City         string `gorm:"type:varchar(20);not null"`
	District     string `gorm:"type:varchar(20);not null"`
	Address      string `gorm:"type:varchar(100);not null"`
	SignerName   string `gorm:"type:varchar(30);not null"`
	SignerMobile string `gorm:"type:varchar(20);not null"`
}

// GetAddressList 某用户的收货地址列表。
func (s *UserOpServer) GetAddressList(ctx context.Context, req *proto.AddressListRequest) (*proto.AddressListResponse, error) {
	var addrs []Address
	result := s.db.Where("user_id = ?", req.UserId).Find(&addrs)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "查询地址失败: %v", result.Error)
	}
	rsp := &proto.AddressListResponse{
		Total: int32(result.RowsAffected),
		Data:  make([]*proto.AddressInfoResponse, 0, result.RowsAffected),
	}
	for _, a := range addrs {
		rsp.Data = append(rsp.Data, addressModelToResp(&a))
	}
	return rsp, nil
}

// CreateAddress 新建地址（强制忽略入参 id）。
func (s *UserOpServer) CreateAddress(ctx context.Context, req *proto.AddressRequest) (*proto.AddressInfoResponse, error) {
	a := Address{
		UserID: req.UserId, Province: req.Province, City: req.City, District: req.District,
		Address: req.Address, SignerName: req.SignerName, SignerMobile: req.SignerMobile,
	}
	if err := s.db.Create(&a).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "创建地址失败: %v", err)
	}
	return addressModelToResp(&a), nil
}

// UpdateAddress 按 id+userId 更新（非本人 NotFound，不泄露存在性）。
func (s *UserOpServer) UpdateAddress(ctx context.Context, req *proto.AddressRequest) (*emptypb.Empty, error) {
	result := s.db.Model(&Address{}).Where("id = ? AND user_id = ?", req.Id, req.UserId).Updates(map[string]any{
		"province":      req.Province,
		"city":          req.City,
		"district":      req.District,
		"address":       req.Address,
		"signer_name":   req.SignerName,
		"signer_mobile": req.SignerMobile,
	})
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "更新地址失败: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "地址不存在")
	}
	return &emptypb.Empty{}, nil
}

// DeleteAddress 按 id+userId 删除（非本人 NotFound）。
func (s *UserOpServer) DeleteAddress(ctx context.Context, req *proto.DeleteAddressRequest) (*emptypb.Empty, error) {
	result := s.db.Where("id = ? AND user_id = ?", req.Id, req.UserId).Delete(&Address{})
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "删除地址失败: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "地址不存在")
	}
	return &emptypb.Empty{}, nil
}

func addressModelToResp(a *Address) *proto.AddressInfoResponse {
	return &proto.AddressInfoResponse{
		Id: a.ID, UserId: a.UserID, Province: a.Province, City: a.City,
		District: a.District, Address: a.Address,
		SignerName: a.SignerName, SignerMobile: a.SignerMobile,
	}
}
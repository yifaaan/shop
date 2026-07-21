package main

import (
	"context"

	basemodel "shop/pkg/model"
	"shop/pkg/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Brands struct {
	basemodel.BaseModel
	Name string `gorm:"type:varchar(50);not null"`
	Logo string `gorm:"type:varchar(200)"`
}

// BrandList 品牌列表（分页）
func (s *GoodsServer) BrandList(ctx context.Context, req *proto.BrandFilterRequest) (*proto.BrandListResponse, error) {
	var brands []Brands
	result := s.db.Scopes(Paginate(int(req.Pages), int(req.PagePerNums))).Find(&brands)
	if result.Error != nil {
		return nil, result.Error
	}

	rsp := &proto.BrandListResponse{
		Total: int32(result.RowsAffected),
		Data:  make([]*proto.BrandInfoResponse, 0, result.RowsAffected),
	}
	for i := range brands {
		rsp.Data = append(rsp.Data, BrandModelToResponse(&brands[i]))
	}
	return rsp, nil
}

// CreateBrand 新建品牌
func (s *GoodsServer) CreateBrand(ctx context.Context, req *proto.BrandRequest) (*proto.BrandInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateBrand not implemented")
}

// DeleteBrand 删除品牌
func (s *GoodsServer) DeleteBrand(ctx context.Context, req *proto.BrandRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteBrand not implemented")
}

// UpdateBrand 更新品牌
func (s *GoodsServer) UpdateBrand(ctx context.Context, req *proto.BrandRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateBrand not implemented")
}

func BrandModelToResponse(brand *Brands) *proto.BrandInfoResponse {
	return &proto.BrandInfoResponse{
		Id:   brand.ID,
		Name: brand.Name,
		Logo: brand.Logo,
	}
}

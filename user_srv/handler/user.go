package handler

import (
	"context"
	"shop/user_srv/global"
	"shop/user_srv/model"
	"shop/user_srv/proto"

	"gorm.io/gorm"
)

type UserServer struct {
	proto.UnimplementedUserServer
}

var _ proto.UserServer = (*UserServer)(nil)

// 获取用户列表
func (s *UserServer) GetUserList(ctx context.Context, req *proto.PageInfoRequest) (*proto.UserListResponse, error) {
	var users []model.User
	// 统计总数
	resp := &proto.UserListResponse{}
	global.DB.Model(&model.User{}).Count(&resp.Total)

	resp.Data = make([]*proto.UserInfoResponse, 0, len(users))

	global.DB.Scopes(Paginate(int(req.PageNumber), int(req.PageSize))).Find(&users)
	for _, u := range users {
		resp.Data = append(resp.Data, ModelToResponse(u))
	}
	return resp, nil
}

// GetUserByMobile 通过手机号获取用户
func (s *UserServer) GetUserByMobile(context.Context, *proto.MobileRequest) (*proto.UserInfoResponse, error) {
}

// GetUserByID 通过ID获取用户
func (s *UserServer) GetUserByID(context.Context, *proto.IDRequest) (*proto.UserInfoResponse, error) {
}
func (s *UserServer) CreateUser(context.Context, *proto.CreateUserInfoRequest) (*proto.UserInfoResponse, error) {
}
func (s *UserServer) UpdateUser(context.Context, *proto.UpdateUserInfoRequest) (*proto.Empty, error) {
}
func (s *UserServer) CheckPassword(context.Context, *proto.CheckPasswordInfoRequest) (*proto.CheckPasswordResponse, error) {
}

// Paginate 分页
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 10
		}
		if pageSize > 100 {
			pageSize = 100
		}
		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// ModelToResponse 模型转换,将model.User转换为proto.UserInfoResponse
func ModelToResponse(user model.User) *proto.UserInfoResponse {
	userInfoResp := proto.UserInfoResponse{
		Id:       user.ID,
		Mobile:   user.Mobile,
		Password: user.Password,
		NickName: user.NickName,
		Gender:   user.Gender,
		Role:     int32(user.Role),
	}
	if user.Birthday != nil {
		userInfoResp.Birthday = uint64(user.Birthday.Unix())
	}
	return &userInfoResp
}

package handler

import (
	"context"
	"shop/pkg/proto"
	"shop/services/user_srv/global"
	"shop/services/user_srv/model"

	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

type UserServer struct{}

// GetUserList 获取用户列表
func (s *UserServer) GetUserList(ctx context.Context, req *proto.PageInfo) (*proto.UserListResponse, error) {
	var users []model.User
	result := global.DB.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	rsp := &proto.UserListResponse{}
	rsp.Total = int32(result.RowsAffected)
	rsp.Data = make([]*proto.UserInfoResponse, 0, rsp.Total)
	global.DB.Scopes(Paginate(int(req.Pn), int(req.PSize))).Find(&users)
	for _, user := range users {
		userInfo := ModelToResponse(&user)
		rsp.Data = append(rsp.Data, userInfo)
	}

	return rsp, nil
}

func (s *UserServer) GetUserByMobile(context.Context, *proto.MobileRequest) (*proto.UserInfoResponse, error) {
	return nil, nil
}

func (s *UserServer) GetUserById(context.Context, *proto.IdRequest) (*proto.UserInfoResponse, error) {
	return nil, nil
}

func (s *UserServer) CreateUser(context.Context, *proto.CreateUserInfo) (*proto.UserInfoResponse, error) {
	return nil, nil
}

func (s *UserServer) UpdateUser(context.Context, *proto.UpdateUserInfo) (*emptypb.Empty, error) {
	return nil, nil
}

func (s *UserServer) CheckPassWord(context.Context, *proto.PasswordCheckInfo) (*proto.CheckResponse, error) {
	return nil, nil
}

// Paginate 分页
func Paginate(page, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page <= 0 {
			page = 1
		}
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func ModelToResponse(user *model.User) *proto.UserInfoResponse {
	userInfo := &proto.UserInfoResponse{
		Id:       user.ID,
		Password: user.Password,
		NickName: user.NickName,
		Gender:   user.Gender,
		Role:     int32(user.Role),
	}
	if user.Birthday != nil {
		userInfo.BirthDay = uint64(user.Birthday.Unix())
	}
	return userInfo
}

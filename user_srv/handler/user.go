package handler

import (
	"context"
	"shop/user_srv/proto"
)

type UserServer struct {
	proto.UnimplementedUserServer
}

func (s *UserServer) GetUserList(context.Context, *proto.PageInfoRequest) (*proto.UserListResponse, error) {

}

func (s *UserServer) GetUserByMobile(context.Context, *proto.MobileRequest) (*proto.UserInfoResponse, error) {
}
func (s *UserServer) GetUserByID(context.Context, *proto.IDRequest) (*proto.UserInfoResponse, error) {
}
func (s *UserServer) CreateUser(context.Context, *proto.CreateUserInfoRequest) (*proto.UserInfoResponse, error) {
}
func (s *UserServer) UpdateUser(context.Context, *proto.UpdateUserInfoRequest) (*proto.Empty, error) {
}
func (s *UserServer) CheckPassword(context.Context, *proto.CheckPasswordInfoRequest) (*proto.CheckPasswordResponse, error) {
}

package handler

import (
	"context"
	"crypto/sha512"
	"fmt"
	"shop/good_srv/global"
	"shop/good_srv/model"
	"shop/good_srv/proto"
	"strings"
	"time"

	"github.com/anaskhan96/go-password-encoder"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	if err := global.DB.WithContext(ctx).Model(&model.User{}).Count(&resp.Total).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "统计用户总数失败: %v", err)
	}

	resp.Data = make([]*proto.UserInfoResponse, 0, len(users))

	if err := global.DB.WithContext(ctx).Scopes(Paginate(int(req.PageNumber), int(req.PageSize))).Find(&users).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "查询用户列表失败: %v", err)
	}
	for _, u := range users {
		resp.Data = append(resp.Data, ModelToResponse(u))
	}
	return resp, nil
}

// GetUserByMobile 通过手机号获取用户
func (s *UserServer) GetUserByMobile(ctx context.Context, req *proto.MobileRequest) (*proto.UserInfoResponse, error) {
	var user model.User
	result := global.DB.WithContext(ctx).Where(&model.User{Mobile: req.Mobile}).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "用户不存在")
		}
		return nil, status.Errorf(codes.Internal, "查询用户失败: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}
	return ModelToResponse(user), nil
}

// GetUserByID 通过ID获取用户
func (s *UserServer) GetUserByID(ctx context.Context, req *proto.IDRequest) (*proto.UserInfoResponse, error) {
	var user model.User
	result := global.DB.WithContext(ctx).First(&user, req.Id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "用户不存在")
		}
		return nil, status.Errorf(codes.Internal, "查询用户失败: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}
	return ModelToResponse(user), nil
}

// CreateUser 创建用户
func (s *UserServer) CreateUser(ctx context.Context, req *proto.CreateUserInfoRequest) (*proto.UserInfoResponse, error) {

	var user model.User
	result := global.DB.WithContext(ctx).Where(&model.User{Mobile: req.Mobile}).First(&user)
	if result.Error == nil && result.RowsAffected == 1 {
		return nil, status.Errorf(codes.AlreadyExists, "用户已存在")
	}
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return nil, status.Errorf(codes.Internal, "查询用户失败: %v", result.Error)
	}
	user.NickName = req.NickName
	user.Mobile = req.Mobile

	// 加密
	options := &password.Options{SaltLen: 16, Iterations: 100, KeyLen: 32, HashFunction: sha512.New}
	salt, encodedPwd := password.Encode(req.Password, options)
	user.Password = fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd)

	result = global.DB.WithContext(ctx).Create(&user)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "创建用户失败: %v", result.Error)
	}
	return ModelToResponse(user), nil
}

// UpdateUser 更新用户
func (s *UserServer) UpdateUser(ctx context.Context, req *proto.UpdateUserInfoRequest) (*proto.Empty, error) {
	var user model.User
	result := global.DB.WithContext(ctx).First(&user, req.Id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, status.Errorf(codes.NotFound, "用户不存在")
		}
		return nil, status.Errorf(codes.Internal, "查询用户失败: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}
	birthday := time.Unix(int64(req.Birthday), 0)
	user.NickName = req.NickName
	user.Gender = req.Gender
	user.Birthday = &birthday
	result = global.DB.WithContext(ctx).Save(&user)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "更新用户失败: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.Internal, "未更新任何记录")
	}
	return &proto.Empty{}, nil
}

// CheckPassword 检查密码
func (s *UserServer) CheckPassword(ctx context.Context, req *proto.CheckPasswordInfoRequest) (*proto.CheckPasswordResponse, error) {
	options := &password.Options{SaltLen: 16, Iterations: 100, KeyLen: 32, HashFunction: sha512.New}
	// 密码格式: $pbkdf2-sha512$salt$encodedPwd
	pwd := strings.Split(req.EncryptedPassword, "$")
	if len(pwd) != 4 {
		return &proto.CheckPasswordResponse{Success: false}, nil
	}
	salt := pwd[2]
	encodedPwd := pwd[3]
	check := password.Verify(req.Password, salt, encodedPwd, options)
	return &proto.CheckPasswordResponse{Success: check}, nil
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

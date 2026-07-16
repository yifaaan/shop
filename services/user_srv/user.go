package main

import (
	"context"
	"crypto/sha512"
	"errors"
	"fmt"
	"strings"
	"time"

	basemodel "shop/pkg/model"
	"shop/pkg/proto"

	"github.com/anaskhan96/go-password-encoder"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

// User is the persisted user record owned by the user service.
type User struct {
	basemodel.BaseModel
	Mobile   string     `gorm:"index:idx_mobile,unique;type:varchar(11);not null"`
	Password string     `gorm:"type:varchar(100);not null"`
	NickName string     `gorm:"type:varchar(20)"`
	Birthday *time.Time `gorm:"type:datetime"` // 指针允许字段为 null
	Gender   string     `gorm:"column:gender;default:male;type:varchar(6) comment 'female 女, male男'"`
	Role     int        `gorm:"column:role;default:1;type:int comment '1表示用户, 2表示管理员'"`
}

// UserServer implements proto.UserServer over an injected *gorm.DB.
type UserServer struct {
	proto.UnimplementedUserServer
	db *gorm.DB
}

// passwordOptions is the shared PBKDF2-SHA512 encoder config for password hashing.
var passwordOptions = &password.Options{
	SaltLen:      10,
	Iterations:   100,
	KeyLen:       16,
	HashFunction: sha512.New,
}

// NewUserServer wires a UserServer to its data store.
func NewUserServer(db *gorm.DB) *UserServer {
	return &UserServer{db: db}
}

// GetUserList 获取用户列表
func (s *UserServer) GetUserList(ctx context.Context, req *proto.PageInfo) (*proto.UserListResponse, error) {
	var users []User
	result := s.db.Scopes(Paginate(int(req.Pn), int(req.PSize))).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	rsp := &proto.UserListResponse{}
	rsp.Total = int32(result.RowsAffected)
	rsp.Data = make([]*proto.UserInfoResponse, 0, rsp.Total)

	for _, user := range users {
		rsp.Data = append(rsp.Data, ModelToResponse(&user))
	}

	return rsp, nil
}

// GetUserByMobile 通过手机号查询用户
func (s *UserServer) GetUserByMobile(ctx context.Context, req *proto.MobileRequest) (*proto.UserInfoResponse, error) {
	var user User
	result := s.db.Where(&User{Mobile: req.Mobile}).First(&user)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}
	if result.Error != nil {
		return nil, result.Error
	}

	return ModelToResponse(&user), nil
}

func (s *UserServer) GetUserById(ctx context.Context, req *proto.IdRequest) (*proto.UserInfoResponse, error) {
	var user User
	result := s.db.First(&user, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}
	if result.Error != nil {
		return nil, result.Error
	}

	return ModelToResponse(&user), nil
}

func (s *UserServer) CreateUser(ctx context.Context, req *proto.CreateUserInfo) (*proto.UserInfoResponse, error) {
	var user User
	result := s.db.Where(&User{Mobile: req.Mobile}).First(&user)
	if result.RowsAffected == 1 {
		return nil, status.Errorf(codes.AlreadyExists, "用户已存在")
	}
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}

	user.Mobile = req.Mobile
	user.NickName = req.NickName

	// 密码加密
	salt, encodedPwd := password.Encode(req.Password, passwordOptions)
	newPassword := fmt.Sprintf("$pbkdf2-sha512$%s$%s", salt, encodedPwd)
	user.Password = newPassword

	result = s.db.Create(&user)
	if result.Error != nil {
		return nil, status.Error(codes.Internal, result.Error.Error())
	}

	return ModelToResponse(&user), nil
}

func (s *UserServer) UpdateUser(ctx context.Context, req *proto.UpdateUserInfo) (*emptypb.Empty, error) {
	var user User
	result := s.db.First(&user, req.Id)
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "用户不存在")
	}
	if result.Error != nil {
		return nil, result.Error
	}

	birthDay := time.Unix(int64(req.BirthDay), 0)
	user.NickName = req.NickName
	user.Birthday = &birthDay
	user.Gender = req.Gender

	result = s.db.Save(&user)
	if result.Error != nil {
		return nil, status.Error(codes.Internal, result.Error.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *UserServer) CheckPassWord(ctx context.Context, req *proto.PasswordCheckInfo) (*proto.CheckResponse, error) {
	parts := strings.Split(req.EncryptedPassword, "$")
	check := password.Verify(req.Password, parts[2], parts[3], passwordOptions)
	return &proto.CheckResponse{Success: check}, nil
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

func ModelToResponse(user *User) *proto.UserInfoResponse {
	userInfo := &proto.UserInfoResponse{
		Id:       user.ID,
		Password: user.Password,
		NickName: user.NickName,
		Gender:   user.Gender,
		Role:     int32(user.Role),
		Mobile:   user.Mobile,
	}
	if user.Birthday != nil {
		userInfo.BirthDay = uint64(user.Birthday.Unix())
	}
	return userInfo
}

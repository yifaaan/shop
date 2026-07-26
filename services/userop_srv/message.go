package main

import (
	"context"

	basemodel "shop/pkg/model"
	"shop/pkg/proto"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Message 留言：type 1留言 2投诉 3询问 4售后 5求购。
type Message struct {
	basemodel.BaseModel
	UserID  int32  `gorm:"index;not null"`
	Subject string `gorm:"type:varchar(100);not null"`
	Message string `gorm:"type:varchar(500);not null"`
	Type    int32  `gorm:"type:int;not null"`
	File    string `gorm:"type:varchar(200)"`
}

// GetMessageList 某用户的留言列表。
func (s *UserOpServer) GetMessageList(ctx context.Context, req *proto.MessageListRequest) (*proto.MessageListResponse, error) {
	var msgs []Message
	result := s.db.Where("user_id = ?", req.UserId).Find(&msgs)
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "查询留言失败: %v", result.Error)
	}
	rsp := &proto.MessageListResponse{
		Total: int32(result.RowsAffected),
		Data:  make([]*proto.MessageInfoResponse, 0, result.RowsAffected),
	}
	for _, m := range msgs {
		rsp.Data = append(rsp.Data, messageModelToResp(&m))
	}
	return rsp, nil
}

// CreateMessage 新建留言
func (s *UserOpServer) CreateMessage(ctx context.Context, req *proto.MessageRequest) (*proto.MessageInfoResponse, error) {
	if req.Type < 1 || req.Type > 5 {
		return nil, status.Errorf(codes.InvalidArgument, "非法留言类型: %d", req.Type)
	}
	m := Message{
		UserID: req.UserId, Subject: req.Subject, Message: req.Message,
		Type: req.Type, File: req.File,
	}
	if err := s.db.Create(&m).Error; err != nil {
		return nil, status.Errorf(codes.Internal, "创建留言失败: %v", err)
	}
	return messageModelToResp(&m), nil
}

// DeleteMessage 按 id+userId 删除
func (s *UserOpServer) DeleteMessage(ctx context.Context, req *proto.DeleteMessageRequest) (*emptypb.Empty, error) {
	result := s.db.Where("id = ? AND user_id = ?", req.Id, req.UserId).Delete(&Message{})
	if result.Error != nil {
		return nil, status.Errorf(codes.Internal, "删除留言失败: %v", result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, status.Errorf(codes.NotFound, "留言不存在")
	}
	return &emptypb.Empty{}, nil
}

func messageModelToResp(m *Message) *proto.MessageInfoResponse {
	return &proto.MessageInfoResponse{
		Id: m.ID, UserId: m.UserID, Subject: m.Subject, Message: m.Message,
		Type: m.Type, File: m.File, AddTime: m.CreatedAt.Unix(),
	}
}

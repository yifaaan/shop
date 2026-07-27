package web

import (
	"net/http"
	"strconv"

	"shop/pkg/proto"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HandleGrpcErrorToHttp maps a gRPC status error onto an HTTP response.
func HandleGrpcErrorToHttp(err error, ctx *gin.Context) {
	if err == nil {
		return
	}
	if e, ok := status.FromError(err); ok {
		switch e.Code() {
		case codes.NotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"msg": e.Message()})
		case codes.Internal:
			ctx.JSON(http.StatusInternalServerError, gin.H{"msg": e.Message()})
		case codes.InvalidArgument:
			ctx.JSON(http.StatusBadRequest, gin.H{"msg": e.Message()})
		case codes.AlreadyExists:
			ctx.JSON(http.StatusConflict, gin.H{"msg": e.Message()})
		case codes.PermissionDenied:
			ctx.JSON(http.StatusForbidden, gin.H{"msg": e.Message()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"msg": e.Message()})
		}
		return
	}
	ctx.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
}

func handleValidatorError(err error, ctx *gin.Context) {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}
	ctx.JSON(http.StatusBadRequest, gin.H{"msg": removeTopStruct(errs.Translate(translator))})
}

// ==================== 收藏 Handlers ====================

// UserFavList 当前用户的收藏列表，返回 snake_case JSON，
// 并通过 goods_srv 批量补全商品名/图/价。
func (s *Server) UserFavList(ctx *gin.Context) {
	rsp, err := s.useropSrv.GetUserFavList(ctx.Request.Context(), &proto.UserFavListRequest{
		UserId: currentUserID(ctx),
	})
	if err != nil {
		s.log.Error("UserFavList grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}

	// 批量拉取商品信息（名/封面图/本店价）补全收藏项
	goodsMap := map[int32]*proto.GoodsInfoResponse{}
	if len(rsp.Data) > 0 {
		ids := make([]int32, 0, len(rsp.Data))
		for _, f := range rsp.Data {
			ids = append(ids, f.GoodsId)
		}
		if gRsp, gErr := s.goodsSrv.BatchGetGoods(ctx.Request.Context(), &proto.BatchGoodsIdInfo{Id: ids}); gErr == nil {
			for _, g := range gRsp.Data {
				goodsMap[g.Id] = g
			}
		} else {
			s.log.Error("UserFavList BatchGetGoods error: ", gErr)
		}
	}

	out := &userFavListResponse{
		Total: rsp.Total,
		Data:  make([]*userFavResponse, 0, len(rsp.Data)),
	}
	for _, f := range rsp.Data {
		out.Data = append(out.Data, userFavToResponse(f, goodsMap[f.GoodsId]))
	}
	ctx.JSON(http.StatusOK, out)
}

// UserFavDetail 查询是否已收藏某商品（未收藏 404）。
func (s *Server) UserFavDetail(ctx *gin.Context) {
	goodsId, err := strconv.Atoi(ctx.Param("goods_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid goods_id"})
		return
	}
	rsp, err := s.useropSrv.GetUserFav(ctx.Request.Context(), &proto.UserFavRequest{
		UserId:  currentUserID(ctx),
		GoodsId: int32(goodsId),
	})
	if err != nil {
		s.log.Error("UserFavDetail grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, userFavToResponse(rsp, nil))
}

// CreateUserFav 收藏商品（重复收藏 409）。
func (s *Server) CreateUserFav(ctx *gin.Context) {
	var form UserFavForm
	if err := ctx.ShouldBind(&form); err != nil {
		handleValidatorError(err, ctx)
		return
	}
	rsp, err := s.useropSrv.CreateUserFav(ctx.Request.Context(), &proto.UserFavRequest{
		UserId:  currentUserID(ctx),
		GoodsId: form.GoodsId,
	})
	if err != nil {
		s.log.Error("CreateUserFav grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, userFavToResponse(rsp, nil))
}

// DeleteUserFav 取消收藏（按 goodsId，未收藏 404）。
func (s *Server) DeleteUserFav(ctx *gin.Context) {
	goodsId, err := strconv.Atoi(ctx.Param("goods_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid goods_id"})
		return
	}
	_, err = s.useropSrv.DeleteUserFav(ctx.Request.Context(), &proto.UserFavRequest{
		UserId:  currentUserID(ctx),
		GoodsId: int32(goodsId),
	})
	if err != nil {
		s.log.Error("DeleteUserFav grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "取消收藏成功"})
}

// ==================== 地址 Handlers ====================

// AddressList 当前用户的收货地址列表。
func (s *Server) AddressList(ctx *gin.Context) {
	rsp, err := s.useropSrv.GetAddressList(ctx.Request.Context(), &proto.AddressListRequest{
		UserId: currentUserID(ctx),
	})
	if err != nil {
		s.log.Error("AddressList grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	out := &addressListResponse{
		Total: rsp.Total,
		Data:  make([]*addressResponse, 0, len(rsp.Data)),
	}
	for _, a := range rsp.Data {
		out.Data = append(out.Data, addressToResponse(a))
	}
	ctx.JSON(http.StatusOK, out)
}

// CreateAddress 新建收货地址。
func (s *Server) CreateAddress(ctx *gin.Context) {
	var form AddressForm
	if err := ctx.ShouldBind(&form); err != nil {
		handleValidatorError(err, ctx)
		return
	}
	rsp, err := s.useropSrv.CreateAddress(ctx.Request.Context(), &proto.AddressRequest{
		UserId:       currentUserID(ctx),
		Province:     form.Province,
		City:         form.City,
		District:     form.District,
		Address:      form.Address,
		SignerName:   form.SignerName,
		SignerMobile: form.SignerMobile,
	})
	if err != nil {
		s.log.Error("CreateAddress grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, addressToResponse(rsp))
}

// UpdateAddress 更新收货地址（id 由 URL 提供，非本人 404）。
func (s *Server) UpdateAddress(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
		return
	}
	var form AddressForm
	if err := ctx.ShouldBind(&form); err != nil {
		handleValidatorError(err, ctx)
		return
	}
	_, err = s.useropSrv.UpdateAddress(ctx.Request.Context(), &proto.AddressRequest{
		Id:           int32(id),
		UserId:       currentUserID(ctx),
		Province:     form.Province,
		City:         form.City,
		District:     form.District,
		Address:      form.Address,
		SignerName:   form.SignerName,
		SignerMobile: form.SignerMobile,
	})
	if err != nil {
		s.log.Error("UpdateAddress grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// DeleteAddress 删除收货地址（非本人 404）。
func (s *Server) DeleteAddress(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
		return
	}
	_, err = s.useropSrv.DeleteAddress(ctx.Request.Context(), &proto.DeleteAddressRequest{
		Id:     int32(id),
		UserId: currentUserID(ctx),
	})
	if err != nil {
		s.log.Error("DeleteAddress grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}

// ==================== 留言 Handlers ====================

// MessageList 当前用户的留言列表。
func (s *Server) MessageList(ctx *gin.Context) {
	rsp, err := s.useropSrv.GetMessageList(ctx.Request.Context(), &proto.MessageListRequest{
		UserId: currentUserID(ctx),
	})
	if err != nil {
		s.log.Error("MessageList grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	out := &messageListResponse{
		Total: rsp.Total,
		Data:  make([]*messageResponse, 0, len(rsp.Data)),
	}
	for _, m := range rsp.Data {
		out.Data = append(out.Data, messageToResponse(m))
	}
	ctx.JSON(http.StatusOK, out)
}

// CreateMessage 新建留言（type 必须 ∈ 1..5）。
func (s *Server) CreateMessage(ctx *gin.Context) {
	var form CreateMessageForm
	if err := ctx.ShouldBind(&form); err != nil {
		handleValidatorError(err, ctx)
		return
	}
	rsp, err := s.useropSrv.CreateMessage(ctx.Request.Context(), &proto.MessageRequest{
		UserId:  currentUserID(ctx),
		Subject: form.Subject,
		Message: form.Message,
		Type:    form.Type,
		File:    form.File,
	})
	if err != nil {
		s.log.Error("CreateMessage grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, messageToResponse(rsp))
}

// DeleteMessage 删除留言（非本人 404）。
func (s *Server) DeleteMessage(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid id"})
		return
	}
	_, err = s.useropSrv.DeleteMessage(ctx.Request.Context(), &proto.DeleteMessageRequest{
		Id:     int32(id),
		UserId: currentUserID(ctx),
	})
	if err != nil {
		s.log.Error("DeleteMessage grpc error: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}
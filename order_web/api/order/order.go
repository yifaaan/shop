package order

import (
	"net/http"
	"shop/order_web/api"
	"shop/order_web/form"
	"shop/order_web/global"
	"shop/order_web/proto"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// List 获取订单列表
func List(ctx *gin.Context) {
	resp, err := global.OrderSrvClient.OrderList(ctx.Request.Context(), &proto.OrderFilterRequest{
		UserId: 1,
	})
	if err != nil {
		zap.S().Errorw("[List] 获取订单列表失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func New(ctx *gin.Context) {

}

func Detail(ctx *gin.Context) {

}

func Delete(ctx *gin.Context) {
	// 解析参数
	id, _ := strconv.Atoi(ctx.Param("id"))
	if id == 0 {
		ctx.Status(http.StatusNotFound)
		return
	}

	zap.S().Info("I'm here")
	// 调用good rpc delete服务
	_, err := global.GoodSrvClient.DeleteGood(ctx.Request.Context(), &proto.DeleteGoodInfo{Id: int32(id)})
	if err != nil {
		zap.S().Errorw("[Delete] 删除【商品】失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}

	ctx.Status(http.StatusOK)
}

func Stock(ctx *gin.Context) {
	// 解析参数
	id, _ := strconv.Atoi(ctx.Param("id"))
	if id == 0 {
		ctx.Status(http.StatusNotFound)
		return
	}
	// TODO:商品库存
	ctx.Status(http.StatusOK)
}

// UpdateStatus 更新商品状态
func UpdateStatus(ctx *gin.Context) {
	// 解析参数
	id, _ := strconv.Atoi(ctx.Param("id"))
	if id == 0 {
		ctx.Status(http.StatusNotFound)
		return
	}

	goodStatusForm := form.GoodStatusForm{}
	if err := ctx.ShouldBind(&goodStatusForm); err != nil {
		zap.S().Errorw("[UpdateStatus] 绑定【商品状态】失败", "msg", err.Error())
		api.HandleValidatorError(ctx, err)
		return
	}
	if _, err := global.GoodSrvClient.UpdateGood(ctx.Request.Context(), &proto.CreateGoodInfo{
		Id:     int32(id),
		IsNew:  goodStatusForm.IsNew,
		IsHot:  goodStatusForm.IsHot,
		OnSale: goodStatusForm.OnSale,
	}); err != nil {
		zap.S().Errorw("[UpdateStatus] 更新【商品状态】失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "更新【商品状态】成功"})
}

// Update 更新商品全部信息
func Update(ctx *gin.Context) {
	// 解析参数
	id, _ := strconv.Atoi(ctx.Param("id"))
	if id == 0 {
		ctx.Status(http.StatusNotFound)
		return
	}

	goodForm := form.GoodForm{}
	if err := ctx.ShouldBind(&goodForm); err != nil {
		zap.S().Errorw("[Update] 绑定【商品】失败", "msg", err.Error())
		api.HandleValidatorError(ctx, err)
		return
	}
	if _, err := global.GoodSrvClient.UpdateGood(ctx.Request.Context(), &proto.CreateGoodInfo{
		Id:             int32(id),
		Name:           goodForm.Name,
		GoodSn:         goodForm.GoodSn,
		Stocks:         goodForm.Stocks,
		CategoryId:     goodForm.CategoryId,
		BrandId:        goodForm.BrandId,
		MarketPrice:    goodForm.MarketPrice,
		ShopPrice:      goodForm.ShopPrice,
		GoodBrief:      goodForm.GoodBrief,
		Images:         goodForm.Images,
		DescImages:     goodForm.DescImages,
		GoodDesc:       goodForm.GoodDesc,
		ShipFree:       goodForm.ShipFree,
		GoodFrontImage: goodForm.FrontImage,
	}); err != nil {
		zap.S().Errorw("[Update] 更新【商品】失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"msg": "更新【商品】成功"})
}

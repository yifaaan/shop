package shop_cart

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

func List(ctx *gin.Context) {
	userID := ctx.GetUint("userID")
	resp, err := global.OrderSrvClient.CartItemList(ctx, &proto.UserInfo{Id: int32(userID)})
	if err != nil {
		zap.S().Errorw("[List] 获取购物车商品列表失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	if resp.Total == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"total": 0,
		})
		return
	}
	// 所有商品id
	goodIds := make([]int32, 0, resp.Total)
	for _, item := range resp.Data {
		goodIds = append(goodIds, item.GoodsId)
	}
	// 批量获取商品信息
	goodInfos, err := global.GoodSrvClient.BatchGetGood(ctx, &proto.BatchGoodIdInfo{Id: goodIds})
	if err != nil {
		zap.S().Errorw("[List] 批量获取商品信息失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}

	goods := make([]map[string]any, 0, goodInfos.Total)
	for _, item := range resp.Data {
		for _, good := range goodInfos.Data {
			if good.Id == item.GoodsId {
				goods = append(goods, map[string]any{
					"id":         item.Id,      // 购物车项id
					"good_id":    item.GoodsId, // 商品id
					"good_name":  good.Name,
					"good_image": good.GoodFrontImage,
					"good_price": good.ShopPrice,
					"nums":       item.Nums,
					"checked":    item.Checked,
				})
			}
		}
	}
	ctx.JSON(http.StatusOK, gin.H{
		"total": resp.Total,
		"data":  goods,
	})
}

func New(ctx *gin.Context) {
	var itemForm form.ShopCartItemForm
	if err := ctx.ShouldBind(&itemForm); err != nil {
		zap.S().Errorw("[New] 绑定【购物车商品】失败", "msg", err.Error())
		api.HandleValidatorError(ctx, err)
		return
	}
	// 查询商品是否存在
	good, err := global.GoodSrvClient.GetGoodDetail(ctx.Request.Context(), &proto.GoodInfoRequest{Id: itemForm.GoodId})
	if err != nil {
		zap.S().Errorw("[New] 查询商品是否存在失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	if good.Id == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "商品不存在"})
		return
	}

	// 查询商品库存
	inv, err := global.InvSrvClient.InvDetail(ctx.Request.Context(), &proto.GoodInvInfo{GoodId: itemForm.GoodId})
	if err != nil {
		zap.S().Errorw("[New] 查询商品库存失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	if inv.Nums < itemForm.Nums {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "商品库存不足"})
		return
	}
	// 添加购物车商品
	userID := ctx.GetUint("userID")
	resp, err := global.OrderSrvClient.CreateCartItem(ctx.Request.Context(), &proto.CartItemRequest{
		UserId:  int32(userID),
		GoodsId: itemForm.GoodId,
		Nums:    itemForm.Nums,
	})
	if err != nil {
		zap.S().Errorw("[New] 添加购物车商品失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
	}
	ctx.JSON(http.StatusOK, gin.H{"id": resp.Id})
}

func Delete(ctx *gin.Context) {
	id, _ := strconv.Atoi(ctx.Param("id"))
	if id == 0 {
		ctx.Status(http.StatusNotFound)
		return
	}
	_, err := global.OrderSrvClient.DeleteCartItem(ctx.Request.Context(), &proto.CartItemRequest{Id: int32(id)})
	if err != nil {
		zap.S().Errorw("[Delete] 删除购物车商品失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	ctx.Status(http.StatusOK)
}

func Update(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "更新购物车商品成功",
	})
}

package banner

import (
	"net/http"
	"shop/good_web/api"
	"shop/good_web/form"
	"shop/good_web/global"
	"shop/good_web/proto"
	"strconv"

	"github.com/gin-gonic/gin"
)

func List(ctx *gin.Context) {
	resp, err := global.GoodSrvClient.BannerList(ctx.Request.Context(), &proto.Empty{})
	if err != nil {
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}

	banners := make([]map[string]any, 0)
	for _, b := range resp.Data {
		banners = append(banners, map[string]any{
			"id":    b.Id,
			"index": b.Index,
			"image": b.Image,
			"url":   b.Url,
		})
	}

	ctx.JSON(http.StatusOK, banners)
}

func New(ctx *gin.Context) {
	bannerForm := form.BannerForm{}
	if err := ctx.ShouldBind(&bannerForm); err != nil {
		api.HandleValidatorError(ctx, err)
		return
	}
	resp, err := global.GoodSrvClient.CreateBanner(ctx.Request.Context(), &proto.BannerRequest{
		Index: int32(bannerForm.Index),
		Image: bannerForm.Image,
		Url:   bannerForm.Url,
	})
	if err != nil {
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	c := map[string]any{
		"id":    resp.Id,
		"index": resp.Index,
		"image": resp.Image,
		"url":   resp.Url,
	}
	ctx.JSON(http.StatusOK, c)
}

func Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		api.HandleValidatorError(ctx, err)
		return
	}
	_, err = global.GoodSrvClient.DeleteBanner(ctx.Request.Context(), &proto.BannerRequest{Id: int32(id)})
	if err != nil {
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, map[string]string{"msg": "删除成功"})
}

func Update(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		api.HandleValidatorError(ctx, err)
		return
	}
	bannerForm := form.BannerForm{}
	if err := ctx.ShouldBind(&bannerForm); err != nil {
		api.HandleValidatorError(ctx, err)
		return
	}
	_, err = global.GoodSrvClient.UpdateBanner(ctx.Request.Context(), &proto.BannerRequest{
		Id:    int32(id),
		Index: int32(bannerForm.Index),
		Image: bannerForm.Image,
		Url:   bannerForm.Url,
	})
	if err != nil {
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, map[string]string{"msg": "更新成功"})
}

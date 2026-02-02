package brand

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
	pages, _ := strconv.Atoi(ctx.DefaultQuery("pn", "1"))
	pagePerNums, _ := strconv.Atoi(ctx.DefaultQuery("psize", "10"))

	resp, err := global.GoodSrvClient.BrandList(ctx.Request.Context(), &proto.BrandFilterRequest{
		Pages:       int32(pages),
		PagePerNums: int32(pagePerNums),
	})
	if err != nil {
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}

	brands := make([]map[string]any, 0)
	for _, b := range resp.Data {
		brands = append(brands, map[string]any{
			"id":   b.Id,
			"name": b.Name,
			"logo": b.Logo,
		})
	}

	ctx.JSON(http.StatusOK, brands)
}

func New(ctx *gin.Context) {
	brandForm := form.BrandForm{}
	if err := ctx.ShouldBind(&brandForm); err != nil {
		api.HandleValidatorError(ctx, err)
		return
	}
	resp, err := global.GoodSrvClient.CreateBrand(ctx.Request.Context(), &proto.BrandRequest{
		Name: brandForm.Name,
		Logo: brandForm.Logo,
	})
	if err != nil {
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	c := map[string]any{
		"id":   resp.Id,
		"name": resp.Name,
		"logo": resp.Logo,
	}
	ctx.JSON(http.StatusOK, c)
}

func Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		api.HandleValidatorError(ctx, err)
		return
	}
	_, err = global.GoodSrvClient.DeleteBrand(ctx.Request.Context(), &proto.BrandRequest{Id: int32(id)})
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
	brandForm := form.BrandForm{}
	if err := ctx.ShouldBind(&brandForm); err != nil {
		api.HandleValidatorError(ctx, err)
		return
	}
	_, err = global.GoodSrvClient.UpdateBrand(ctx.Request.Context(), &proto.BrandRequest{
		Id:   int32(id),
		Name: brandForm.Name,
		Logo: brandForm.Logo,
	})
	if err != nil {
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, map[string]string{"msg": "更新成功"})
}

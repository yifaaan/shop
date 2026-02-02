package brand

import (
	"net/http"
	"shop/good_web/api"
	"shop/good_web/form"
	"shop/good_web/global"
	"shop/good_web/proto"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func BrandList(ctx *gin.Context) {
	pages, _ := strconv.Atoi(ctx.DefaultQuery("pn", "1"))
	pagePerNums, _ := strconv.Atoi(ctx.DefaultQuery("psize", "10"))

	resp, err := global.GoodSrvClient.BrandList(ctx.Request.Context(), &proto.BrandFilterRequest{
		Pages:       int32(pages),
		PagePerNums: int32(pagePerNums),
	})
	if err != nil {
		zap.S().Errorw("[BrandList] 获取【品牌】列表失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}

	brands := make([]map[string]any, 0, len(resp.Data))
	for _, b := range resp.Data {
		brands = append(brands, map[string]any{
			"id":   b.Id,
			"name": b.Name,
			"logo": b.Logo,
		})
	}

	ctx.JSON(http.StatusOK, brands)
}

func NewBrand(ctx *gin.Context) {
	brandForm := form.BrandForm{}
	if err := ctx.ShouldBind(&brandForm); err != nil {
		zap.S().Errorw("[NewBrand] 绑定【品牌】失败", "msg", err.Error())
		api.HandleValidatorError(ctx, err)
		return
	}
	resp, err := global.GoodSrvClient.CreateBrand(ctx.Request.Context(), &proto.BrandRequest{
		Name: brandForm.Name,
		Logo: brandForm.Logo,
	})
	if err != nil {
		zap.S().Errorw("[NewBrand] 创建【品牌】失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id":   resp.Id,
		"name": resp.Name,
		"logo": resp.Logo,
	})
}

func DeleteBrand(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		zap.S().Errorw("[DeleteBrand] 解析【品牌ID】失败", "msg", err.Error())
		api.HandleValidatorError(ctx, err)
		return
	}
	_, err = global.GoodSrvClient.DeleteBrand(ctx.Request.Context(), &proto.BrandRequest{Id: int32(id)})
	if err != nil {
		zap.S().Errorw("[DeleteBrand] 删除【品牌】失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "删除成功",
	})
}

func UpdateBrand(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		api.HandleValidatorError(ctx, err)
		return
	}
	brandForm := form.BrandForm{}
	if err := ctx.ShouldBind(&brandForm); err != nil {
		zap.S().Errorw("[UpdateBrand] 绑定【品牌】失败", "msg", err.Error())
		api.HandleValidatorError(ctx, err)
		return
	}
	_, err = global.GoodSrvClient.UpdateBrand(ctx.Request.Context(), &proto.BrandRequest{
		Id:   int32(id),
		Name: brandForm.Name,
		Logo: brandForm.Logo,
	})
	if err != nil {
		zap.S().Errorw("[UpdateBrand] 更新【品牌】失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "更新成功",
	})
}

// GetCategoryBrand 通过分类ID获取品牌列表
func GetCategoryBrand(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		zap.S().Errorw("[GetCategoryBrand] 解析【分类ID】失败", "msg", err.Error())
		api.HandleValidatorError(ctx, err)
		return
	}
	resp, err := global.GoodSrvClient.GetCategoryBrandList(ctx.Request.Context(), &proto.CategoryInfoRequest{Id: int32(id)})
	if err != nil {
		zap.S().Errorw("[GetCategoryBrand] 获取【品牌分类关联】列表失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	brands := make([]map[string]any, 0, len(resp.Data))
	for _, b := range resp.Data {
		brands = append(brands, map[string]any{
			"id":   b.Id,
			"name": b.Name,
			"logo": b.Logo,
		})
	}
	ctx.JSON(http.StatusOK, brands)
}

// GetCategoryBrandList 获取品牌分类关联列表
func GetCategoryBrandList(ctx *gin.Context) {
	pages, _ := strconv.Atoi(ctx.DefaultQuery("pn", "1"))
	pagePerNums, _ := strconv.Atoi(ctx.DefaultQuery("psize", "10"))
	resp, err := global.GoodSrvClient.CategoryBrandList(ctx.Request.Context(), &proto.CategoryBrandFilterRequest{
		Pages:       int32(pages),
		PagePerNums: int32(pagePerNums),
	})
	if err != nil {
		zap.S().Errorw("[GetCategoryBrandList] 获取【品牌分类关联】列表失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}

	categoryBrands := make([]map[string]any, 0, len(resp.Data))
	for _, cb := range resp.Data {
		categoryBrands = append(categoryBrands, map[string]any{
			"id": cb.Id,
			"category": map[string]any{
				"id":   cb.Category.Id,
				"name": cb.Category.Name,
			},
			"brand": map[string]any{
				"id":   cb.Brand.Id,
				"name": cb.Brand.Name,
				"logo": cb.Brand.Logo,
			},
		})
	}

	ctx.JSON(http.StatusOK, gin.H{
		"total": resp.Total, // 表的总记录数
		"data":  categoryBrands,
	})
}

func NewCategoryBrand(ctx *gin.Context) {
	categoryBrandForm := form.CategoryBrandForm{}
	if err := ctx.ShouldBind(&categoryBrandForm); err != nil {
		zap.S().Errorw("[NewCategoryBrand] 绑定【品牌分类关联】失败", "msg", err.Error())
		api.HandleValidatorError(ctx, err)
		return
	}
	resp, err := global.GoodSrvClient.CreateCategoryBrand(ctx.Request.Context(), &proto.CategoryBrandRequest{
		CategoryId: categoryBrandForm.CategoryId,
		BrandId:    categoryBrandForm.BrandId,
	})
	if err != nil {
		zap.S().Errorw("[NewCategoryBrand] 创建【品牌分类关联】失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id": resp.Id,
	})
}

func UpdateCategoryBrand(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		api.HandleValidatorError(ctx, err)
		return
	}
	categoryBrandForm := form.CategoryBrandForm{}
	if err := ctx.ShouldBind(&categoryBrandForm); err != nil {
		zap.S().Errorw("[UpdateCategoryBrand] 绑定【品牌分类关联】失败", "msg", err.Error())
		api.HandleValidatorError(ctx, err)
		return
	}
	_, err = global.GoodSrvClient.UpdateCategoryBrand(ctx.Request.Context(), &proto.CategoryBrandRequest{
		Id:         int32(id),
		CategoryId: categoryBrandForm.CategoryId,
		BrandId:    categoryBrandForm.BrandId,
	})
	if err != nil {
		zap.S().Errorw("[UpdateCategoryBrand] 更新【品牌分类关联】失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "更新成功",
	})
}

func DeleteCategoryBrand(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		zap.S().Errorw("[DeleteCategoryBrand] 解析【品牌分类关联ID】失败", "msg", err.Error())
		ctx.Status(http.StatusNotFound)
		return
	}
	_, err = global.GoodSrvClient.DeleteCategoryBrand(ctx.Request.Context(), &proto.CategoryBrandRequest{Id: int32(id)})
	if err != nil {
		zap.S().Errorw("[DeleteCategoryBrand] 删除【品牌分类关联】失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "删除成功",
	})
}

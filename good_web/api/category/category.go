package category

import (
	"encoding/json"
	"net/http"
	"shop/good_web/api"
	"shop/good_web/form"
	"shop/good_web/global"
	"shop/good_web/proto"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// List 获取所有分类列表
func List(ctx *gin.Context) {
	resp, err := global.GoodSrvClient.GetAllCategorysList(ctx.Request.Context(), &proto.Empty{})
	if err != nil {
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	// resp中的JsonData是json字符串，需要反序列化
	data := make([]map[string]any, 0)
	err = json.Unmarshal([]byte(resp.JsonData), &data)
	if err != nil {
		zap.S().Errorw("[List] 反序列化【分类列表】失败", "msg", err.Error())
	}
	ctx.JSON(http.StatusOK, data)
}

func Detail(ctx *gin.Context) {
	// category id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		zap.S().Errorw("[Detail] 解析【分类ID】失败", "msg", err.Error())
		ctx.Status(http.StatusNotFound)
		return
	}

	resp, err := global.GoodSrvClient.GetSubCategory(ctx.Request.Context(), &proto.CategoryListRequest{
		Id: int32(id),
	})
	if err != nil {
		zap.S().Errorw("[Detail] 获取【分类】失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}

	subCategorys := make([]map[string]any, len(resp.SubCategorys))
	for _, v := range resp.SubCategorys {
		subCategorys = append(subCategorys, map[string]any{
			"id":              v.Id,
			"name":            v.Name,
			"level":           v.Level,
			"is_tab":          v.IsTab,
			"parent_category": v.ParentCategory,
		})
	}
	c := map[string]any{
		"id":              resp.Info.Id,
		"name":            resp.Info.Name,
		"level":           resp.Info.Level,
		"is_tab":          resp.Info.IsTab,
		"parent_category": resp.Info.ParentCategory,
		"sub_categorys":   subCategorys,
	}
	ctx.JSON(http.StatusOK, c)
}

func New(ctx *gin.Context) {
	categoryForm := form.CategoryForm{}
	if err := ctx.ShouldBind(&categoryForm); err != nil {
		zap.S().Errorw("[New] 绑定【分类】失败", "msg", err.Error())
		api.HandleValidatorError(ctx, err)
		return
	}

	resp, err := global.GoodSrvClient.CreateCategory(ctx.Request.Context(), &proto.CategoryInfoRequest{
		Name:           categoryForm.Name,
		Level:          categoryForm.Level,
		ParentCategory: categoryForm.ParentCategory,
		IsTab:          categoryForm.IsTab,
	})
	if err != nil {
		zap.S().Errorw("[New] 创建【分类】失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}

	c := map[string]any{
		"id":     resp.Id,
		"name":   resp.Name,
		"level":  resp.Level,
		"is_tab": resp.IsTab,
		"parent": resp.ParentCategory,
	}
	ctx.JSON(http.StatusOK, c)
}

func Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		zap.S().Errorw("[Delete] 解析【分类ID】失败", "msg", err.Error())
		ctx.Status(http.StatusNotFound)
		return
	}
	_, err = global.GoodSrvClient.DeleteCategory(ctx.Request.Context(), &proto.DeleteCategoryRequest{
		Id: int32(id),
	})

	if err != nil {
		zap.S().Errorw("[Delete] 删除【分类】失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, map[string]string{"msg": "删除成功"})
}

func Update(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		zap.S().Errorw("[Update] 解析【分类ID】失败", "msg", err.Error())
		ctx.Status(http.StatusNotFound)
		return
	}
	categoryForm := form.UpdateCategoryForm{}
	if err := ctx.ShouldBind(&categoryForm); err != nil {
		zap.S().Errorw("[Update] 绑定【分类】失败", "msg", err.Error())
		api.HandleValidatorError(ctx, err)
		return
	}
	_, err = global.GoodSrvClient.UpdateCategory(ctx.Request.Context(), &proto.CategoryInfoRequest{
		Id:    int32(id),
		Name:  categoryForm.Name,
		IsTab: categoryForm.IsTab,
	})
	if err != nil {
		zap.S().Errorw("[Update] 更新【分类】失败", "msg", err.Error())
		api.HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	ctx.JSON(http.StatusOK, map[string]string{"msg": "更新成功"})
}

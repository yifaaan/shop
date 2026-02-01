package category

import (
	"encoding/json"
	"net/http"
	"shop/good_web/api"
	"shop/good_web/global"
	"shop/good_web/proto"

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

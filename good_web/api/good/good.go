package good

import (
	"net/http"
	"shop/good_web/global"
	"shop/good_web/proto"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// removeTopStruct 移除结构体名称,只保留字段名
func removeTopStruct(fields map[string]string) map[string]string {
	res := map[string]string{}
	for field, err := range fields {
		res[field[strings.Index(field, ".")+1:]] = err
	}
	return res
}

// HandleGrpcErrorToHttpError 将grpc错误转换为http错误
func HandleGrpcErrorToHttpError(err error, c *gin.Context) {
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				c.JSON(http.StatusNotFound, gin.H{"msg": st.Message()})
			case codes.Internal:
				c.JSON(http.StatusInternalServerError, gin.H{"msg": st.Message()})
			case codes.InvalidArgument:
				c.JSON(http.StatusBadRequest, gin.H{"msg": st.Message()})
			case codes.Unauthenticated:
				c.JSON(http.StatusUnauthorized, gin.H{"msg": st.Message()})
			case codes.PermissionDenied:
				c.JSON(http.StatusForbidden, gin.H{"msg": st.Message()})
			case codes.AlreadyExists:
				c.JSON(http.StatusConflict, gin.H{"msg": st.Message()})
			case codes.ResourceExhausted:
				c.JSON(http.StatusTooManyRequests, gin.H{"msg": st.Message()})
			case codes.FailedPrecondition:
				c.JSON(http.StatusPreconditionFailed, gin.H{"msg": st.Message()})
			case codes.Aborted:
				c.JSON(http.StatusConflict, gin.H{"msg": st.Message()})
			case codes.OutOfRange:
				c.JSON(http.StatusBadRequest, gin.H{"msg": st.Message()})
			case codes.Unimplemented:
				c.JSON(http.StatusNotImplemented, gin.H{"msg": st.Message()})
			case codes.Unavailable:
				c.JSON(http.StatusServiceUnavailable, gin.H{"msg": st.Message()})
			case codes.DeadlineExceeded:
				c.JSON(http.StatusRequestTimeout, gin.H{"msg": st.Message()})
			case codes.Canceled:
				c.JSON(http.StatusRequestTimeout, gin.H{"msg": st.Message()})
			case codes.Unknown:
				c.JSON(http.StatusInternalServerError, gin.H{"msg": st.Message()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"msg": st.Message()})
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		}
	}
}

// HandleValidatorError 处理表单验证错误
func HandleValidatorError(ctx *gin.Context, err error) {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		zap.S().Errorw("[HandleValidatorError] 转换为validator.ValidationErrors失败", "msg", err.Error())
		ctx.JSON(http.StatusOK, gin.H{"msg": err.Error()})
		return
	}
	ctx.JSON(http.StatusBadRequest, gin.H{"err": removeTopStruct(errs.Translate(global.Trans))})
}

func List(ctx *gin.Context) {
	// 过滤参数解析
	request := &proto.GoodFilterRequest{}
	priceMin, _ := strconv.Atoi(ctx.DefaultQuery("pmin", "0"))
	request.PriceMax = int32(priceMin)
	priceMax, _ := strconv.Atoi(ctx.DefaultQuery("pmax", "0"))
	request.PriceMax = int32(priceMax)
	isHot, _ := strconv.ParseBool(ctx.DefaultQuery("ih", "false"))
	request.IsHot = isHot
	isNew, _ := strconv.ParseBool(ctx.DefaultQuery("in", "false"))
	request.IsNew = isNew
	isTab, _ := strconv.ParseBool(ctx.DefaultQuery("it", "false"))
	request.IsTab = isTab
	categoryId, _ := strconv.Atoi(ctx.DefaultQuery("c", "0"))
	request.TopCategory = int32(categoryId)
	pages, _ := strconv.Atoi(ctx.DefaultQuery("pn", "0"))
	request.Pages = int32(pages)
	perNums, _ := strconv.Atoi(ctx.DefaultQuery("pnum", "0"))
	request.PagePerNums = int32(perNums)
	keywords := ctx.DefaultQuery("q", "")
	request.KeyWords = keywords
	brandId, _ := strconv.Atoi(ctx.DefaultQuery("b", "0"))
	request.Brand = int32(brandId)

	// 商品rpc服务
	resp, err := global.GoodSrvClient.GoodList(ctx.Request.Context(), request)
	if err != nil {
		zap.S().Errorw("[List] 查询【商品列表】失败", "msg", err.Error())
		HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	// 返回响应
	goodList := make([]any, 0, len(resp.Data))
	for _, g := range resp.Data {
		goodList = append(goodList, map[string]any{
			"id":          g.Id,
			"name":        g.Name,
			"goods_brief": g.GoodBrief,
			"desc":        g.GoodDesc,
			"ship_free":   g.ShipFree,
			"images":      g.Images,
			"desc_images": g.DescImages,
			"front_image": g.GoodFrontImage,
			"shop_price":  g.ShopPrice,

			"category": map[string]any{
				"id":   g.Category.Id,
				"name": g.Category.Name,
			},
			"brand": map[string]any{
				"id":   g.Brand.Id,
				"name": g.Brand.Name,
				"logo": g.Brand.Logo,
			},
			"is_hot":  g.IsHot,
			"is_new":  g.IsNew,
			"on_sale": g.OnSale,
		})
	}
	m := map[string]any{
		"total": resp.Total,
		"data":  goodList,
	}
	ctx.JSON(http.StatusOK, m)
}

package api

import (
	"net/http"
	"shop/shop_api/user_web/form"
	"shop/shop_api/user_web/global"
	"shop/shop_api/user_web/global/response"
	"shop/shop_api/user_web/proto"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

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

func GetUserList(ctx *gin.Context) {
	conn, err := grpc.NewClient("127.0.0.1:8080", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zap.S().Errorw("[GetUserList] 连接【用户服务】失败", "msg", err.Error())
		HandleGrpcErrorToHttpError(err, ctx)
		return
	}

	// 创建rpc用户服务客户端
	userSrvClient := proto.NewUserClient(conn)

	// 解析参数
	pn := ctx.DefaultQuery("pn", "1")
	pnInt, _ := strconv.Atoi(pn)
	pSize := ctx.DefaultQuery("psize", "10")
	pSizeInt, _ := strconv.Atoi(pSize)
	// 调用用户服务
	resp, err := userSrvClient.GetUserList(ctx.Request.Context(), &proto.PageInfoRequest{
		PageNumber: uint32(pnInt),
		PageSize:   uint32(pSizeInt),
	})
	if err != nil {
		zap.S().Errorw("[GetUserList] 查询 【用户列表】失败", "msg", err.Error())
		HandleGrpcErrorToHttpError(err, ctx)
		return
	}

	// 构造响应
	zap.S().Debugf("获取用户列表: %v", resp)
	result := make([]response.UserResponse, 0, len(resp.Data))
	for _, u := range resp.Data {
		result = append(result, response.UserResponse{
			Id:       u.Id,
			NickName: u.NickName,
			Mobile:   u.Mobile,
			Gender:   u.Gender,
			Birthday: response.JsonTime(time.Unix(int64(u.Birthday), 0)),
			Role:     u.Role,
		})
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

func Login(ctx *gin.Context) {
	// 表单验证
	loginForm := form.LoginForm{}
	if err := ctx.ShouldBind(&loginForm); err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ctx.JSON(http.StatusOK, gin.H{"msg": err.Error()})
			return
		}
		zap.S().Errorw("[Login] 表单验证失败", "msg", err.Error())
		// 翻译错误
		ctx.JSON(http.StatusBadRequest, gin.H{"err": removeTopStruct(errs.Translate(global.Trans))})
		return
	}
}

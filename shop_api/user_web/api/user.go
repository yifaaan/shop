package api

import (
	"net/http"
	"shop/shop_api/user_web/proto"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

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

	// 创建用户服务客户端
	userSrvClient := proto.NewUserClient(conn)

	resp, err := userSrvClient.GetUserList(ctx.Request.Context(), &proto.PageInfoRequest{
		PageNumber: 1,
		PageSize:   10,
	})
	if err != nil {
		zap.S().Errorw("[GetUserList] 查询 【用户列表】失败", "msg", err.Error())
		HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	zap.S().Debugf("获取用户列表: %v", resp)
	result := make([]any, 0, len(resp.Data))
	for _, u := range resp.Data {
		data := make(map[string]any)
		data["id"] = u.Id
		data["nick_name"] = u.NickName
		data["mobile"] = u.Mobile
		data["gender"] = u.Gender
		data["birthday"] = u.Birthday
		data["role"] = u.Role
		result = append(result, data)
	}
	ctx.JSON(http.StatusOK, gin.H{"data": result})
}

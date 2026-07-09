package api

import (
	"fmt"
	"net/http"
	"shop/pkg/proto"
	"shop/services/user_web/global"
	"shop/services/user_web/global/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func HandleGrpcErrorToHttp(err error, ctx *gin.Context) {
	if err == nil {
		return
	}

	if e, ok := status.FromError(err); ok {
		switch e.Code() {
		case codes.NotFound:
			ctx.JSON(http.StatusNotFound, gin.H{"msg": e.Message()})
		case codes.Internal:
			ctx.JSON(http.StatusInternalServerError, gin.H{"msg": e.Message()})
		case codes.InvalidArgument:
			ctx.JSON(http.StatusBadRequest, gin.H{"msg": e.Message()})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{"msg": e.Message()})
		}
	}
}

func GetUserList(ctx *gin.Context) {
	zap.S().Debug("GetUserList called")

	// Connect to the user service via gRPC
	userConn, err := grpc.NewClient(fmt.Sprintf("%s:%d", global.ServerConfig.UserSrv.Host, global.ServerConfig.UserSrv.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zap.S().Error("failed to connect to user service: ", err.Error())
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	defer userConn.Close()

	userSrvClient := proto.NewUserClient(userConn)

	pn := ctx.DefaultQuery("pn", "0")
	psize := ctx.DefaultQuery("psize", "10")
	pnInt, _ := strconv.Atoi(pn)
	psizeInt, _ := strconv.Atoi(psize)
	srvRsp, err := userSrvClient.GetUserList(ctx.Request.Context(), &proto.PageInfo{
		Pn:    uint32(pnInt),
		PSize: uint32(psizeInt),
	})
	if err != nil {
		zap.S().Error("failed to get user list: ", err.Error())
		HandleGrpcErrorToHttp(err, ctx)
		return
	}

	result := make([]response.UserResponse, 0, srvRsp.Total)
	for _, val := range srvRsp.Data {
		user := response.UserResponse{
			Id:       val.Id,
			NickName: val.NickName,
			Mobile:   val.Mobile,
			Gender:   val.Gender,
			Birthday: response.JsonTime(time.Unix(int64(val.BirthDay), 0)),
		}
		result = append(result, user)
	}
	ctx.JSON(http.StatusOK, result)
}

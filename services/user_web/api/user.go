package api

import (
	"fmt"
	"net/http"
	"shop/pkg/proto"

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
	ip := "127.0.0.1"
	port := 50051
	zap.S().Debug("GetUserList called")

	// Connect to the user service via gRPC
	userConn, err := grpc.NewClient(fmt.Sprintf("%s:%d", ip, port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zap.S().Error("failed to connect to user service: ", err.Error())
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	defer userConn.Close()

	userSrvClient := proto.NewUserClient(userConn)

	srvRsp, err := userSrvClient.GetUserList(ctx.Request.Context(), &proto.PageInfo{
		Pn:    0,
		PSize: 0,
	})
	if err != nil {
		zap.S().Error("failed to get user list: ", err.Error())
		HandleGrpcErrorToHttp(err, ctx)
		return
	}

	result := make([]any, 0, srvRsp.Total)
	for _, val := range srvRsp.Data {
		data := make(map[string]any)
		data["id"] = val.Id
		data["name"] = val.NickName
		data["mobile"] = val.Mobile
		data["birthday"] = val.BirthDay
		result = append(result, data)
	}
	ctx.JSON(http.StatusOK, result)
}

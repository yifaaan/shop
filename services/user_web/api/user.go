package api

import (
	"fmt"
	"net/http"
	"shop/pkg/model"
	"shop/pkg/proto"
	"shop/services/user_web/forms"
	"shop/services/user_web/global"
	"shop/services/user_web/global/response"
	"shop/services/user_web/middlewares"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
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

func removeTopStruct(fields map[string]string) map[string]string {
	res := map[string]string{}
	for field, err := range fields {
		res[field[strings.Index(field, ".")+1:]] = err
	}
	return res
}

func handleValidatorError(err error, ctx *gin.Context) {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}
	ctx.JSON(http.StatusBadRequest, gin.H{"msg": removeTopStruct(errs.Translate(global.Translator))})
}
func GetUserList(ctx *gin.Context) {
	uid := ctx.GetUint("userId")
	zap.S().Debug("GetUserList called, userId: ", uid)

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

func PasswordLogin(ctx *gin.Context) {
	var loginForm forms.PasswordLoginForm
	if err := ctx.ShouldBind(&loginForm); err != nil {
		zap.S().Error("PasswordLogin form binding error: ", err.Error())
		handleValidatorError(err, ctx)
		return
	}

	// Verify captcha
	if !store.Verify(loginForm.CaptchaId, loginForm.Captcha, true) {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid captcha"})
		return
	}

	userConn, err := grpc.NewClient(fmt.Sprintf("%s:%d", global.ServerConfig.UserSrv.Host, global.ServerConfig.UserSrv.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zap.S().Error("failed to connect to user service: ", err.Error())
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	defer userConn.Close()

	userSrvClient := proto.NewUserClient(userConn)

	rsp, err := userSrvClient.GetUserByMobile(ctx.Request.Context(), &proto.MobileRequest{
		Mobile: loginForm.Mobile,
	})
	if err != nil {
		zap.S().Error("failed to get user by mobile: ", err.Error())
		HandleGrpcErrorToHttp(err, ctx)
		return
	}

	checkRsp, err := userSrvClient.CheckPassWord(ctx.Request.Context(), &proto.PasswordCheckInfo{
		Password:          loginForm.Password,
		EncryptedPassword: rsp.Password,
	})
	if err != nil {
		zap.S().Error("failed to check password: ", err.Error())
		HandleGrpcErrorToHttp(err, ctx)
		return
	}

	if !checkRsp.Success {
		ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "invalid password"})
		return
	}

	// Generate JWT token
	j := middlewares.NewJWT()
	claims := &model.CustomClaims{
		ID:          uint(rsp.Id),
		NickName:    rsp.NickName,
		AuthorityId: uint(rsp.Role),
		StandardClaims: jwt.StandardClaims{
			NotBefore: time.Now().Unix(),            // Token is valid from now
			ExpiresAt: time.Now().Unix() + 60*60*24, // 1 day
			Issuer:    "shop",                       // Issuer
		},
	}

	token, err := j.CreateToken(claims)
	if err != nil {
		zap.S().Error("failed to create token: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "failed to create token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": rsp.Id, "nickname": rsp.NickName, "token": token})
}

func Register(ctx *gin.Context) {
	var registerForm forms.RegisterForm
	if err := ctx.ShouldBind(&registerForm); err != nil {
		zap.S().Error("Register form binding error: ", err.Error())
		handleValidatorError(err, ctx)
		return
	}

	// Verify SMS code
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", global.ServerConfig.Redis.Host, global.ServerConfig.Redis.Port),
		Password: global.ServerConfig.Redis.Password,
		DB:       global.ServerConfig.Redis.DB,
	})
	defer rdb.Close()
	code, err := rdb.Get(ctx.Request.Context(), "sms:code:"+registerForm.Mobile).Result()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "验证码已过期"})
		return
	}
	if code != registerForm.Code {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "验证码错误"})
		return
	}

	// Proceed with registration logic, e.g., call user service to create a new user
	userConn, err := grpc.NewClient(fmt.Sprintf("%s:%d", global.ServerConfig.UserSrv.Host, global.ServerConfig.UserSrv.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zap.S().Error("failed to connect to user service: ", err.Error())
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	defer userConn.Close()

	userSrvClient := proto.NewUserClient(userConn)
	srvRsp, err := userSrvClient.CreateUser(ctx.Request.Context(), &proto.CreateUserInfo{
		Mobile:   registerForm.Mobile,
		Password: registerForm.Password,
		NickName: registerForm.Mobile, // Default nickname is the mobile number
	})
	if err != nil {
		zap.S().Error("failed to create user: ", err.Error())
		HandleGrpcErrorToHttp(err, ctx)
		return
	}

	// Generate JWT token
	j := middlewares.NewJWT()
	claims := &model.CustomClaims{
		ID:          uint(srvRsp.Id),
		NickName:    srvRsp.NickName,
		AuthorityId: uint(srvRsp.Role),
		StandardClaims: jwt.StandardClaims{
			NotBefore: time.Now().Unix(),            // Token is valid from now
			ExpiresAt: time.Now().Unix() + 60*60*24, // 1 day
			Issuer:    "shop",                       // Issuer
		},
	}

	token, err := j.CreateToken(claims)
	if err != nil {
		zap.S().Error("failed to create token: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "failed to create token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"id": srvRsp.Id, "nickname": srvRsp.NickName, "token": token})
}

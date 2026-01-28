package api

import (
	"fmt"
	"net/http"
	"shop/shop_api/user_web/form"
	"shop/shop_api/user_web/global"
	"shop/shop_api/user_web/global/response"
	"shop/shop_api/user_web/middleware"
	"shop/shop_api/user_web/model"
	"shop/shop_api/user_web/proto"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hashicorp/consul/api"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
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

func GetUserList(ctx *gin.Context) {
	// 从注册中心获取user_srv的信息
	// 设置consul信息获取client
	cfg := api.DefaultConfig()
	cfg.Address = fmt.Sprintf("%s:%d", global.ServerConfig.ConsulConfig.Host, global.ServerConfig.ConsulConfig.Port)
	client, err := api.NewClient(cfg)
	if err != nil {
		zap.S().Errorw("[GetUserList] 连接【consul】失败", "msg", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "连接【consul】失败"})
		return
	}

	// 根据服务名查询服务ip+port
	var userSrvAddr string
	var userSrvPort int
	serviceName := global.ServerConfig.UserSrvCfg.Name
	zap.S().Infow("[GetUserList] 正在从Consul查询服务", "service_name", serviceName)

	data, err := client.Agent().ServicesWithFilter(fmt.Sprintf(`Service == "%s"`, serviceName))
	if err != nil {
		zap.S().Errorw("[GetUserList] 获取【用户服务】地址失败", "msg", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "获取【用户服务】地址失败"})
		return
	}

	// 从 Consul 获取服务地址
	if len(data) == 0 {
		zap.S().Errorw("[GetUserList] 从Consul未找到服务", "service", serviceName)
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"msg": fmt.Sprintf("服务【%s】未在Consul中注册", serviceName)})
		return
	}

	// 记录所有找到的服务实例
	zap.S().Infow("[GetUserList] 从Consul找到服务实例", "count", len(data))
	for id, v := range data {
		zap.S().Infow("[GetUserList] 服务实例详情", "service_id", id, "address", v.Address, "port", v.Port, "tags", v.Tags)
	}

	// 获取第一个匹配的服务实例
	found := false
	for _, v := range data {
		userSrvAddr = v.Address
		userSrvPort = v.Port
		zap.S().Infow("[GetUserList] 选择服务实例", "address", userSrvAddr, "port", userSrvPort, "service_id", v.ID)
		found = true
		break
	}

	if !found || userSrvAddr == "" {
		zap.S().Errorw("[GetUserList] Consul中的服务地址无效", "address", userSrvAddr, "port", userSrvPort)
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"msg": "Consul中的服务地址无效"})
		return
	}

	// 与服务建立连接
	userSrvEndpoint := fmt.Sprintf("%s:%d", userSrvAddr, userSrvPort)
	zap.S().Infow("[GetUserList] 正在连接用户服务", "endpoint", userSrvEndpoint)
	conn, err := grpc.NewClient(userSrvEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zap.S().Errorw("[GetUserList] 连接【用户服务】失败", "msg", err.Error())
		HandleGrpcErrorToHttpError(err, ctx)
		return
	}

	userID := ctx.GetUint("userID")
	zap.S().Debugf("GetUserList userID: %d", userID)

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

func LoginPassword(ctx *gin.Context) {
	// 表单验证
	loginForm := form.LoginForm{}
	if err := ctx.ShouldBind(&loginForm); err != nil {
		HandleValidatorError(ctx, err)
		return
	}

	// 验证码
	if !store.Verify(loginForm.CaptchaId, loginForm.Captcha, true) {
		zap.S().Infow("[LoginPassword] 验证码错误", "captchaId", loginForm.CaptchaId, "captcha", loginForm.Captcha)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"captcha": "验证码错误",
		})
		return
	}

	// 连接用户服务
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", global.ServerConfig.UserSrvCfg.Host,
		global.ServerConfig.UserSrvCfg.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zap.S().Errorw("[LoginPassword] 连接【用户服务】失败", "msg", err.Error())
		HandleGrpcErrorToHttpError(err, ctx)
		return
	}
	userSrvClient := proto.NewUserClient(conn)

	// login逻辑
	// 获取user
	if u, err := userSrvClient.GetUserByMobile(ctx.Request.Context(), &proto.MobileRequest{
		Mobile: loginForm.Mobile,
	}); err != nil {
		zap.S().Errorw("[LoginPassword] 查询【用户】失败", "msg", err.Error())
		errs, ok := status.FromError(err)
		if ok {
			switch errs.Code() {
			case codes.NotFound:
				ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "用户不存在"})
				return
			default:
				ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "登录失败"})
			}
		}
		return
	} else {
		// 检查密码
		ok, err := userSrvClient.CheckPassword(ctx.Request.Context(), &proto.CheckPasswordInfoRequest{
			Password:          loginForm.Password,
			EncryptedPassword: u.Password,
		})
		if err != nil {
			zap.S().Errorw("[LoginPassword] 校验密码失败", "msg", err.Error())
			HandleGrpcErrorToHttpError(err, ctx)
			return
		}
		if !ok.Success {
			zap.S().Errorw("[LoginPassword] 密码错误", "msg", "密码错误")
			ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "密码错误"})
			return
		}

		// 登录成功
		// 生成JWT
		j := middleware.NewJWT()
		// 创建claims
		claims := model.CustomClaims{
			ID:          uint(u.Id),
			NickName:    u.NickName,
			AuthorityID: uint(u.Role),
			RegisteredClaims: jwt.RegisteredClaims{
				NotBefore: jwt.NewNumericDate(time.Now()),                                               // iat
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(j.ExpiresAt))), // exp
				Issuer:    j.Issuer,
			},
		}
		// 创建token
		token, err := j.CreateToken(claims)
		if err != nil {
			zap.S().Errorw("[LoginPassword] 创建token失败", "msg", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "创建token失败"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"id":         u.Id,
			"nick_name":  u.NickName,
			"token":      token,
			"expired_at": claims.ExpiresAt.Time.Unix(),
		})
	}

}

func Register(ctx *gin.Context) {
	// 参数验证
	registerForm := form.RegisterForm{}
	if err := ctx.ShouldBind(&registerForm); err != nil {
		zap.S().Errorw("[Register] 参数绑定失败", "msg", err.Error())
		HandleValidatorError(ctx, err)
		return
	}

	// 连接redis
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", global.ServerConfig.RedisConfig.Host, global.ServerConfig.RedisConfig.Port),
	})
	// 获取验证码
	code, err := rdb.Get(ctx.Request.Context(), "sms_code_"+registerForm.Mobile).Result()
	if err != nil {
		zap.S().Errorw("[Register] 获取验证码失败", "msg", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "验证码错误"})
		return
	}
	if code != registerForm.Code {
		zap.S().Errorw("[Register] 验证码错误", "msg", "验证码错误")
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "验证码错误"})
		return
	}

	// 创建rpc连接
	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", global.ServerConfig.UserSrvCfg.Host,
		global.ServerConfig.UserSrvCfg.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		zap.S().Errorw("[Register] 连接【用户服务】失败", "msg", err.Error())
		HandleGrpcErrorToHttpError(err, ctx)
		return
	}

	// 创建用户rpc服务客户端
	userSrvClient := proto.NewUserClient(conn)
	u, err := userSrvClient.CreateUser(ctx.Request.Context(), &proto.CreateUserInfoRequest{
		NickName: registerForm.Mobile,
		Password: registerForm.Password,
		Mobile:   registerForm.Mobile,
	})
	if err != nil {
		zap.S().Errorw("[Register] 创建用户失败", "msg", err.Error())
		HandleGrpcErrorToHttpError(err, ctx)
		return
	}

	// 创建成功，切换到已登录状态
	// 生成JWT
	j := middleware.NewJWT()
	// 创建claims
	claims := model.CustomClaims{
		ID:          uint(u.Id),
		NickName:    u.NickName,
		AuthorityID: uint(u.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			NotBefore: jwt.NewNumericDate(time.Now()),                                               // iat
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * time.Duration(j.ExpiresAt))), // exp
			Issuer:    j.Issuer,
		},
	}
	// 创建token
	token, err := j.CreateToken(claims)
	if err != nil {
		zap.S().Errorw("[LoginPassword] 创建token失败", "msg", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "创建token失败"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id":         u.Id,
		"nick_name":  u.NickName,
		"token":      token,
		"expired_at": claims.ExpiresAt.Time.Unix(),
	})
}

package web

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"shop/pkg/model"
	"shop/pkg/proto"
	"shop/services/user_web/sms"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/mojocn/base64Captcha"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// store is the in-memory captcha store backing GetCaptcha / login verification.
var store = base64Captcha.DefaultMemStore

func (s *Server) GetUserList(ctx *gin.Context) {
	uid := ctx.GetUint("userId")
	s.log.Debug("GetUserList called, userId: ", uid)

	pn := ctx.DefaultQuery("pn", "0")
	psize := ctx.DefaultQuery("psize", "10")
	pnInt, _ := strconv.Atoi(pn)
	psizeInt, _ := strconv.Atoi(psize)

	srvRsp, err := s.userSrv.GetUserList(ctx.Request.Context(), &proto.PageInfo{
		Pn:    uint32(pnInt),
		PSize: uint32(psizeInt),
	})
	if err != nil {
		s.log.Error("failed to get user list: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}

	result := make([]UserResponse, 0, srvRsp.Total)
	for _, val := range srvRsp.Data {
		result = append(result, UserResponse{
			Id:       val.Id,
			NickName: val.NickName,
			Mobile:   val.Mobile,
			Gender:   val.Gender,
			Birthday: JsonTime(time.Unix(int64(val.BirthDay), 0)),
		})
	}
	ctx.JSON(http.StatusOK, result)
}

func (s *Server) PasswordLogin(ctx *gin.Context) {
	var loginForm PasswordLoginForm
	if err := ctx.ShouldBind(&loginForm); err != nil {
		s.log.Error("PasswordLogin form binding error: ", err)
		handleValidatorError(err, ctx)
		return
	}

	// Verify captcha
	if !store.Verify(loginForm.CaptchaId, loginForm.Captcha, true) {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "invalid captcha"})
		return
	}

	rsp, err := s.userSrv.GetUserByMobile(ctx.Request.Context(), &proto.MobileRequest{
		Mobile: loginForm.Mobile,
	})
	if err != nil {
		s.log.Error("failed to get user by mobile: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}

	checkRsp, err := s.userSrv.CheckPassWord(ctx.Request.Context(), &proto.PasswordCheckInfo{
		Password:          loginForm.Password,
		EncryptedPassword: rsp.Password,
	})
	if err != nil {
		s.log.Error("failed to check password: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}

	if !checkRsp.Success {
		ctx.JSON(http.StatusUnauthorized, gin.H{"msg": "invalid password"})
		return
	}

	token, err := s.createToken(rsp.Id, rsp.NickName, rsp.Role)
	if err != nil {
		s.log.Error("failed to create token: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "failed to create token"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"id": rsp.Id, "nickname": rsp.NickName, "token": token})
}

func (s *Server) Register(ctx *gin.Context) {
	var registerForm RegisterForm
	if err := ctx.ShouldBind(&registerForm); err != nil {
		s.log.Error("Register form binding error: ", err)
		handleValidatorError(err, ctx)
		return
	}

	// Verify SMS code
	ok, err := s.sms.VerifyCode(ctx.Request.Context(), registerForm.Mobile, registerForm.Code)
	if errors.Is(err, sms.ErrCodeExpired) {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "验证码已过期"})
		return
	}
	if err != nil {
		s.log.Error("verify sms code error: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "验证码服务暂时不可用"})
		return
	}
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": "验证码错误"})
		return
	}

	srvRsp, err := s.userSrv.CreateUser(ctx.Request.Context(), &proto.CreateUserInfo{
		Mobile:   registerForm.Mobile,
		Password: registerForm.Password,
		NickName: registerForm.Mobile, // Default nickname is the mobile number
	})
	if err != nil {
		s.log.Error("failed to create user: ", err)
		HandleGrpcErrorToHttp(err, ctx)
		return
	}

	token, err := s.createToken(srvRsp.Id, srvRsp.NickName, srvRsp.Role)
	if err != nil {
		s.log.Error("failed to create token: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "failed to create token"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"id": srvRsp.Id, "nickname": srvRsp.NickName, "token": token})
}

func (s *Server) GetCaptcha(ctx *gin.Context) {
	driver := base64Captcha.NewDriverDigit(80, 240, 5, 0.7, 80)
	c := base64Captcha.NewCaptcha(driver, store)
	id, b64s, _, err := c.Generate()
	if err != nil {
		s.log.Error("failed to generate captcha: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"msg": "failed to generate captcha"})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"captchaId": id, "picPath": b64s})
}

// Health is the liveness/readiness endpoint used by Consul.
func (s *Server) Health(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) SendSms(ctx *gin.Context) {
	var form SendSmsForm
	if err := ctx.ShouldBind(&form); err != nil {
		handleValidatorError(err, ctx)
		return
	}

	if err := s.sms.SendCode(ctx.Request.Context(), form.Mobile); err != nil {
		ctx.JSON(smsErrorStatus(err), gin.H{"msg": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"msg": "发送成功"})
}

// createToken mints a 24h JWT for the given user.
func (s *Server) createToken(id int32, nickname string, role int32) (string, error) {
	claims := &model.CustomClaims{
		ID:          uint(id),
		NickName:    nickname,
		AuthorityId: uint(role),
		StandardClaims: jwt.StandardClaims{
			NotBefore: time.Now().Unix(),            // Token is valid from now
			ExpiresAt: time.Now().Unix() + 60*60*24, // 1 day
			Issuer:    "shop",                       // Issuer
		},
	}
	return s.j.CreateToken(claims)
}

// HandleGrpcErrorToHttp maps a gRPC status error onto an HTTP response.
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

func handleValidatorError(err error, ctx *gin.Context) {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}
	ctx.JSON(http.StatusBadRequest, gin.H{"msg": removeTopStruct(errs.Translate(translator))})
}

func removeTopStruct(fields map[string]string) map[string]string {
	res := map[string]string{}
	for field, err := range fields {
		res[field[strings.Index(field, ".")+1:]] = err
	}
	return res
}

func smsErrorStatus(err error) int {
	switch {
	case errors.Is(err, sms.ErrTooFrequent):
		return http.StatusTooManyRequests
	case errors.Is(err, sms.ErrSendFailed), errors.Is(err, sms.ErrNoCode):
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}

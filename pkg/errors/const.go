package errors

import (
	"errors"
	"fmt"
)

var (
	GraphqlObjectIsNull = errors.New("graphql object is null.")
	ServiceAlreadyExist = errors.New("service already registerred.")
	ParamEmpty          = errors.New("param empty.")
	EnvVarNotExist      = errors.New("env variable not exists.")
	TracerIsNull        = errors.New("global tracer is null.")
)

//FGErrorCode All API Errors
type FGErrorCode int

/*
FGErrorCode All API Errors
*/
const (
	FGEBase = 10000
)

//系统错误50000+iota
const (
	FGEInternalError FGErrorCode = 5*FGEBase + iota // value --> 0
	FGEOptimisticLockException
	FGEDBError
	FGECacheError
	FGEHTTPRPCError
	FGEDataParseError
	FGEZKConfigError
	FGENoPermission
	FGEUploadPictureLimit
)

//客户端请求错误 40000+iota
const (
	FGEInvalidUnlockType FGErrorCode = 4*FGEBase + iota // value --> 0

	FGEInvalidRequestParam
	FGEInvalidMobile
	FGECheckSignFail
	FGETrafficControl
	FGESendShortMessage
	FGEOrderNotExist
	FGEOrderNameExist
	FGEUpdateOrder
	FGEInvalidRefreshToken // value --> 1
	FGEInvalidToken
	FGEInvalidUserID
	FGEInvalidVerifyCode
	FGRNotAlowToLogin
	FGEExpiredVerifyCode
	FGEInvalidVersionCode
	FGEInvalidOperation
	FGEInvalidMsgID
	FGEInvalidClientID
	FGESmsFrequence
	FGEVehicleAlreadyBinded
	FGEInvalidVehicleNo
	FGEAdNameError
)

var FGErrorPrefix = "FGError:"
var evtDesc = map[FGErrorCode]string{
	//系统错误 5xx
	FGEInternalError:           "系统内部错误",
	FGEOptimisticLockException: "乐观锁异常",
	FGEDBError:                 "数据库连接或读写异常",
	FGECacheError:              "缓存连接或读写异常",
	FGEDataParseError:          "数据解析失败",
	FGEZKConfigError:           "配置中心配置错误",
	FGENoPermission:            "无操作权限",
	FGEUploadPictureLimit:      "照片上传已达到9张, 无法继续上传",

	//客户端请求错误 4xx
	FGECheckSignFail:        "签名检查不通过",
	FGETrafficControl:       "频控错误",
	FGEInvalidRefreshToken:  "RefreshToken无效或者已过期",
	FGEInvalidToken:         "Token无效或者已过期",
	FGEInvalidVerifyCode:    "验证码无效",
	FGEInvalidMobile:        "无效手机号",
	FGEInvalidUserID:        "无效用户ID",
	FGEInvalidRequestParam:  "请求参数不匹配",
	FGESendShortMessage:     "发送短信失败",
	FGEOrderNotExist:        "订单信息不存在",
	FGEOrderNameExist:       "订单名称已存在，请修改后重新提交",
	FGEUpdateOrder:          "订单状态已更新，请刷新后再试",
	FGRNotAlowToLogin:       "当前用户不允许登录",
	FGEInvalidVersionCode:   "无效版本号",
	FGEInvalidOperation:     "无效操作",
	FGEInvalidMsgID:         "无效消息编号",
	FGEInvalidClientID:      "无效推送消息用户编号",
	FGEExpiredVerifyCode:    "验证码错误或已超时",
	FGESmsFrequence:         "验证码获取太频繁，请稍后再试",
	FGEHTTPRPCError:         "内部微服务调用失败",
	FGEVehicleAlreadyBinded: "该车辆已绑定",
	FGEInvalidVehicleNo:     "无效车牌号",
}

// return all errors
func GetErrors() map[FGErrorCode]string {
	return evtDesc
}

//String FGErrorCode to String
func (f FGErrorCode) String() string {
	if value, ok := evtDesc[f]; ok {
		return value
	}
	return "Unknown"
}

//String FGErrorCode to String
func (f FGErrorCode) Error() string {
	//file, line := WhereAmI(6)
	return fmt.Sprintf("%s%d:%s", FGErrorPrefix, int(f), f.String())
}

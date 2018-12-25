package rpc

import (
	"strings"

	"github.com/microsvs/base/pkg/errors"
)

type FGService int

const FGBService FGService = 8080 // base service port

// 基础服务
const (
	FGSIgnore FGService = FGBService + iota // 忽略
	FGSGateway
	FGSSign
	FGSTraffic
	FGSToken
	FGSUser
	FGSAddress
	FGSImage
)

var serviceMap = map[FGService]string{
	FGSIgnore:  "test",
	FGSGateway: "gateway",
	FGSSign:    "sign",
	FGSTraffic: "traffic",
	FGSToken:   "token",
	FGSUser:    "user",
	FGSAddress: "address",
	FGSImage:   "pictures",
}

func RegisterService(service FGService, serviceName string) error {
	if strings.TrimSpace(serviceName) == "" {
		return errors.ParamEmpty
	}
	if _, ok := serviceMap[service]; ok {
		return errors.ServiceAlreadyExist
	}
	for _, value := range serviceMap {
		if value == serviceName {
			return errors.ServiceAlreadyExist
		}
	}
	serviceMap[service] = serviceName
	return nil
}

func (fg FGService) String() string {
	if value, ok := serviceMap[fg]; ok {
		return value
	}
	return "unknown service name"
}

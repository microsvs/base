package env

import (
	"os"
	"strings"

	"github.com/microsvs/base/pkg/errors"
)

// ENV_NAME表示服务所在server相关环境变量
/*
FREEGO_ZK = "127.0.0.1:2181" // 配置中心地址
FREEGO_NAME = "xxx"       // 服务名称
FREEGO_ENV = "ns-xxx-dev" // 开发环境
FREEGO_LOG = "/var/log/xxx // 日志目录
*/
type ENV_NAME string

const (
	KVConfig    ENV_NAME = "APP_ZK"
	ServiceName ENV_NAME = "APP_NAME"
	ServiceENV  ENV_NAME = "APP_ENV"
	ServiceLog  ENV_NAME = "APP_LOG"
	ServiceVer  ENV_NAME = "APP_VERSION"
	TracerAgent ENV_NAME = "APP_TRACER_AGENT"
)

var (
	defaultLogPath         = "/var/log/app"
	defaultServiceEnv      = "developer"
	defaultDiscoveryConfig = "127.0.0.1:2181"
	defaultServiceVersion  = "v1.0"
	defaultTracerAgent     = "0.0.0.0:6831"
	envMap                 map[ENV_NAME]string
)

func init() {
	var (
		logPathPrefix string
		serviceEnv    string
		discovery     string
		version       string
		tracerAgent   string
	)
	// register service name
	registerMap(ServiceName, initServiceName())

	// register log path prefix
	if logPathPrefix = os.Getenv(string(ServiceLog)); len(logPathPrefix) <= 0 {
		logPathPrefix = defaultLogPath
	}
	registerMap(ServiceLog, logPathPrefix)

	// register service env
	if serviceEnv = os.Getenv(string(ServiceENV)); len(serviceEnv) <= 0 {
		serviceEnv = defaultServiceEnv
	}
	registerMap(ServiceENV, serviceEnv)

	// register discovery configuration
	if discovery = os.Getenv(string(KVConfig)); len(discovery) <= 0 {
		discovery = defaultDiscoveryConfig
	}
	registerMap(KVConfig, discovery)

	if version = os.Getenv(string(ServiceVer)); len(version) <= 0 {
		version = defaultServiceVersion
	}
	registerMap(ServiceVer, version)

	if tracerAgent = os.Getenv(string(TracerAgent)); len(tracerAgent) <= 0 {
		tracerAgent = defaultTracerAgent
	}
	registerMap(TracerAgent, tracerAgent)
}

func initServiceName() string {
	var appname string
	if appname = os.Getenv(string(ServiceName)); len(appname) <= 0 {
		pathes := strings.Split(os.Args[0], "/")
		processName := pathes[len(pathes)-1]
		parts := strings.Split(processName, ".")

		appname = parts[0]
	}
	return appname
}

func registerMap(key ENV_NAME, value string) {
	if envMap == nil {
		envMap = make(map[ENV_NAME]string)
	}
	envMap[key] = value
	return
}

func Get(key ENV_NAME) (string, error) {
	if _, ok := envMap[key]; !ok {
		return "", errors.EnvVarNotExist
	}
	return envMap[key], nil
}

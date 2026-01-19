package logger

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Log 是一个全局变量，以后在任何地方引入这个包就可以直接用
var Log *zap.Logger

func InitLogger() {
	// 获取配置中的日志配置
	logMode := viper.GetString("server.mode")
	
	var config zap.Config
	if logMode == "debug" {
		// 开发模式：日志是彩色的，方便看
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		// 生产模式：日志是 JSON 格式，速度快，方便机器收集
		config = zap.NewProductionConfig()
	}

	// 设置输出位置（默认标准输出）
	config.OutputPaths = []string{"stdout"}
	
	// 构建日志器
	var err error
	Log, err = config.Build()
	if err != nil {
		// 如果日志都起不来，那程序也没法跑了
		panic("❌ 日志初始化失败: " + err.Error())
	}
	
	Log.Info("✅ 日志系统初始化完成", zap.String("env", logMode))
}
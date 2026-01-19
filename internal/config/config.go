package config

import (
	"log"
	"github.com/spf13/viper"
)

// InitConfig 初始化配置，失败直接 panic，不要犹豫
func InitConfig() {
	// 1. 告诉 Viper 我们要读的文件名叫 "config" (不需要 .yaml 后缀)
	viper.SetConfigName("config") 
	
	// 2. 告诉 Viper 文件类型是 yaml
	viper.SetConfigType("yaml")   
	
	// 3. 告诉 Viper 去哪里找文件
	// 我们添加两个路径，防止你在不同目录下运行程序找不到文件
	viper.AddConfigPath("./configs") // 相对路径：当前目录下的 configs
	viper.AddConfigPath("../../configs") // 防止有时你在子目录下测试

	// 4. 开始读取！
	if err := viper.ReadInConfig(); err != nil {
		// 如果读不到，直接炸掉程序。这叫 "Fail Fast"（快速失败）原则
		log.Fatalf("❌ 致命错误：读取配置文件失败: %v \n", err)
	}
	
	log.Println("✅ 配置文件加载成功！")
}
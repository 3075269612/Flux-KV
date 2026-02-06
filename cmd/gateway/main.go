package main

import (
	"Flux-KV/internal/gateway/handler"
	"Flux-KV/internal/gateway/router"
	"Flux-KV/pkg/client"
	"Flux-KV/pkg/discovery"
	"Flux-KV/pkg/logger"
	"Flux-KV/pkg/tracer"
	"context"
	"errors"
	"net/http"
	_ "net/http/pprof"	// 引入 Pprof，自动注册路由
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	// 1. 初始化配置
	viper.SetDefault("server.mode", "debug")        // 默认开发模式
	viper.SetDefault("server.port", "8080")         // 默认端口
	viper.SetDefault("etcd.endpoints", []string{"localhost:2379"})
	viper.SetDefault("rpc.service_name", "kv-service")

	// 2. 初始化日志
	logger.InitLogger()
	// 程序退出前刷新日志缓冲区，防止日志丢失
	defer logger.Log.Sync()

	// 初始化分布式链路追踪
	tp, err := tracer.InitTracer("gateway-service", "localhost:4317")
	if err != nil {
		logger.Log.Error("❌ Failed to init tracer", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			logger.Log.Error("Error shutting down tracer provider", zap.Error(err))
		}
	}()

	// 获取全局 Logger 实例
	log := logger.Log
	log.Info("🚀 Gateway is starting...")

	// 3. 设置 Gin 的运行模式
	gin.SetMode(viper.GetString("server.mode"))

	// Day 17 新增：服务发现与负载均衡链接逻辑
	// A. 连接 Etcd
	etcdEndpoints := viper.GetStringSlice("etcd.endpoints")
	log.Info("🔍 Connecting to Etcd...", zap.Strings("endpoints", etcdEndpoints))

	disco, err := discovery.NewDiscovery(etcdEndpoints)
	if err != nil {
		log.Fatal("❌ Failed to connect to Etcd", zap.Error(err))
	}
	defer disco.Close()	// 退出时关闭 Etcd 连接

	// B. 初始化支持负载均衡的 gRPC Client
	serviceName := viper.GetString("rpc.service_name")
	log.Info("🔗 Initializing KV Client (Load Balanced)...", zap.String("service", serviceName))

	// 注意：这里传入 discovery 实例和服务名，不再是具体的 IP
	kvClient, err := client.NewClient(disco, serviceName)
	if err != nil {
		log.Fatal("❌ Failed to init KV client", zap.Error(err))
	}
	defer func() {
		log.Info("🔌 Closing gRPC client connections...")
		if err := kvClient.Close(); err != nil {
			log.Error("Failed to close gRPC connection", zap.Error(err))
		}
	}()

	// 4. 初始化 Handlers (控制层)
	kvHandler := handler.NewKVHandler(kvClient)
	healthHandler := handler.NewHealthHandler()

	// 5. 初始化 Router (路由层)
	r := router.NewRouter(kvHandler, healthHandler)

	// Day 19 新增
	// 启动 Pprof 监控服务 (独立端口 :6060)
	go func() {
		pprofAddr := "0.0.0.0:6060"
		log.Info("📈 Pprof Debug Server is running", zap.String("addr", "http://localhost:6060/debug/pprof/"))

		// http.ListenAndServe 使用默认的 ServeMux
		if err := http.ListenAndServe(pprofAddr, nil); err != nil {
			log.Error("❌ Pprof Server failed", zap.Error(err))
		}
	}()

	// 6. 配置 HTTP Server
	port := viper.GetString("server.port")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// 7. 启动服务
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("❌ Listen error", zap.Error(err))
		}
	}()
	log.Info("✅ Gateway running", zap.String("port", port))

	// 8. 优雅退出
	quit := make(chan os.Signal, 1)
	// 监听中断信号 (Ctrl+C, Docker stop)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞直到收到信号
	<-quit
	log.Info("⚠️ Shutting down gateway...")

	// 创建一个 5 秒超时的 Context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭服务器，处理完当前的请求
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("❌ Server forced to shutdown", zap.Error(err))
	}

	log.Info("👋 Gateway exited properly")
}

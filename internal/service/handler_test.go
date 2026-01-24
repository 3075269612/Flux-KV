package service

import (
	"Go-AI-KV-System/internal/config"
	"Go-AI-KV-System/internal/core"
	pb "Go-AI-KV-System/proto"
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TestKVServiceFlow 会模拟启动一个服务器，然后创建一个客户端去连接它
func TestKVServiceFlow(t *testing.T) {
	// 1. 启动服务端

	// 1.1 准备基础设施：使用 :0 让系统自动分配空闲端口
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to listen: %v ", err)
	}
	// 获取系统实际分配的端口
	port := lis.Addr().(*net.TCPAddr).Port
	t.Logf("Test Server started on port: %d", port)

	// 1.2 创建 gRPC 服务器，注册 KV 服务
	s := grpc.NewServer()
	db := core.NewMemDB(&config.Config{})
	pb.RegisterKVServiceServer(s, NewKVService(db))

	// 1.3 goroutine 中启动服务（不阻塞测试主线程）
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Errorf("failed to serve: %v", err)
		}
	}()
	defer s.Stop()	// 测试结束自动停止服务，释放资源

	// 2. 启动客户端

	// 2.1 连接到刚才启动的服务端
	addr := lis.Addr().String()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// 2.2 创建客户端存根
	client := pb.NewKVServiceClient(conn)

	// 3. 开始测试
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 3.1 测试 Set
	key, val := "test_key", "hello_grpc"
	_, err = client.Set(ctx, &pb.SetRequest{Key: key, Value: val})
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	t.Log("Set check passed")

	// 3.2 测试 Get
	getResp, err := client.Get(ctx, &pb.GetRequest{Key: key})
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if getResp.Value != val {
		t.Errorf("Get value mismatch: got %v, want %v", getResp.Value, val)
	}
	t.Logf("Get check passed: %v", getResp.Value)

	// 3.3 测试 Del
	_, err = client.Del(ctx, &pb.DelRequest{Key: key})
	if err != nil {
		t.Fatalf("Del failed: %v", err)
	}
	t.Log("Del check passed")

	// 3.4 验证 Del 效果
	getRespAfterDel, err := client.Get(ctx, &pb.GetRequest{Key: key})
	if err != nil {
		t.Fatalf("Get after Del failed: %v", err)
	}
	if getRespAfterDel.Found {
		t.Errorf("Del not effective: key %s still exists", key)
	}
	t.Log("Del effect check passed: key not found")
}
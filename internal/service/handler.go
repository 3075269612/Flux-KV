package service

import (
	"context"
	"Go-AI-KV-System/internal/core"
	pb "Go-AI-KV-System/proto"
)

// 定义服务结构体
type KVService struct {
	// gRPC 的保底实现
	pb.UnimplementedKVServiceServer

	// 持有内存数据库
	db *core.MemDB
}

// 构造函数
func NewKVService(db *core.MemDB) *KVService {
	return &KVService{
		db: db,
	}
}

// 下面是实现 .proto 里定义的三个接口

// 1. 实现 Set
func (s *KVService) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	// 核心逻辑：拿到请求里的 Key, Value，塞给数据库
	s.db.Set(req.Key, req.Value, 0)
	return &pb.SetResponse{
		Success: true,
	}, nil
}

// 2. Get 接口
func (s *KVService) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	// 核心逻辑：去数据库查
	val, found := s.db.Get(req.Key)
	if !found {
		return &pb.GetResponse{
			Value: "",
			Found: false,
		}, nil
	}
	strVal, ok := val.(string)
	if !ok {
		strVal = ""
	}

	return &pb.GetResponse{
		Value: strVal,
		Found: found,
	}, nil
}

// 3. Del 接口
func (s *KVService) Del(ctx context.Context, req *pb.DelRequest) (*pb.DelResponse, error) {
	s.db.Del(req.Key)
	return &pb.DelResponse{
		Success: true,
	}, nil
}
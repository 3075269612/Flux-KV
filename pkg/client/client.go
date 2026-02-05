package client

import (
	pb "Go-AI-KV-System/api/proto"
	"Go-AI-KV-System/pkg/discovery"
	"context"
	"errors"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client 封装了 gRPC 连接池和负载均衡策略
type Client struct {
	mu      sync.RWMutex
	conns   map[string]*grpc.ClientConn   // addr -> 原始连接
	clients map[string]pb.KVServiceClient // addr -> 客户端存根
	addrs   []string	// 节点地址列表（用于轮训索引）

	seq uint64 // 轮询计数器
}

// NewClient 初始化客户端管理器，并开始监听服务节点变化
func NewClient(d *discovery.Discovery, serviceName string) (*Client, error) {
	c := &Client{
		clients: make(map[string]pb.KVServiceClient),
		conns: make(map[string]*grpc.ClientConn),
		addrs: make([]string, 0),
	}
	
	// 启动监听 (回调函数会自动处理现有节点和未来节点的连接建立)
	// 假设 Etcd 中注册的 Key 是 /services/kv-service/uuid
	// 我们监听的前缀就是 /services/kv-service/
	prefix := "/services/" + serviceName + "/"
	err := d.WatchService(prefix, c.addNode, c.removeNode)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Close 关闭底层连接
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for addr, conn := range c.conns {
		conn.Close()
		log.Printf("🔌 [Client] 关闭连接: %s", addr)
	}
	return nil
}

// addNode 节点上线回调：建立连接并加入池子
func (c *Client) addNode(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	addr := value	// Etcd value 存储的是 "ip:port"

	// 防止重复添加
	if _, ok := c.clients[addr]; ok {
		return
	}

	// 建立 gRPC 连接
	conn, err := grpc.NewClient(addr, 
	grpc.WithTransportCredentials(insecure.NewCredentials()), 
	grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		log.Printf("❌ [Client] 连接节点失败 %s: %v", addr, err)
		return
	}

	// 创建存根
	rpcClient := pb.NewKVServiceClient(conn)

	c.clients[addr] = rpcClient
	c.conns[addr] = conn
	c.addrs = append(c.addrs, addr)

	log.Printf("✅ [Client] 节点上线: %s (当前可用: %d)", addr, len(c.addrs))
}

// removeNode 节点下线回调：关闭连接并移出池子
func (c *Client) removeNode(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	addr := value

	// 关闭连接
	if conn, ok := c.conns[addr]; ok {
		conn.Close()
		delete(c.conns, addr)
	}
	delete(c.clients, addr)

	// 从切片中移除地址
	newAddrs := make([]string, 0, len(c.addrs))
	for _, a := range c.addrs {
		if a != addr {
			newAddrs = append(newAddrs, a)
		}
	}
	c.addrs = newAddrs

	log.Printf("❌ [Client] 节点下线: %s (当前可用: %d)", addr, len(c.addrs))
}

// Load Balance 轮询选择一个节点
func (c *Client) lb() (pb.KVServiceClient, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.addrs) == 0 {
		return nil, errors.New("no available kv-service nodes")
	}

	// 核心：原子递增，实现 Round-Robin
	next := atomic.AddUint64(&c.seq, 1)
	index := next % uint64(len(c.addrs))

	targetAddr := c.addrs[index]
	client := c.clients[targetAddr]

	return client, nil
}

// Set 封装 Set 请求
func (c *Client) Set(key, value string) error {
	client, err := c.lb()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
	defer cancel()

	_, err = client.Set(ctx, &pb.SetRequest{Key: key, Value: value})
	return err
}

// Get 封装 Get 请求
func (c *Client) Get(key string) (string, error) {
	client, err := c.lb()
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
	defer cancel()

	resp, err := client.Get(ctx, &pb.GetRequest{Key: key})
	if err != nil {
		return "", err
	}
	return resp.Value, nil
}

// Del 封装 Del 请求
func (c *Client) Del(key string) error {
	client, err := c.lb()
	if err != nil {
		return err
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 2 * time.Second)
	defer cancel()

	_, err = client.Del(ctx, &pb.DelRequest{Key: key})
	return err
}

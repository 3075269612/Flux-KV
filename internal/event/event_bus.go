package event

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// EventType 定义事件类型
type EventType int

const (
	EventSet EventType = iota // 0: 写入/更新
	EventDel                  // 1: 删除
)

// Event 定义了传送带上的盘子里装什么
type Event struct {
	Type  EventType `json:"type"`
	Key   string    `json:"key"`
	Value any       `json:"value"`
}

// EventBus 事件总线核心结构
type EventBus struct {
	// 缓冲通道
	ch           chan Event
	rabbitConn   *amqp.Connection
	rabbitCh     *amqp.Channel
	exchangeName string
	mu           sync.Mutex // 保护日志输出的同步
}

// NewEventBus 初始化传送带
// bufferSize: 缓冲区大小
func NewEventBus(bufferSize int, mqURL string) (*EventBus, error) {
	// 1. 连接 RabbitMQ
	conn, err := amqp.Dial(mqURL)
	if err != nil {
		return nil, err
	}

	// 2. 打开一个 Channel (AMQP 的概念)
	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// 3. 声明交换机（Exchange）
	exchangeName := "flux_kv_events"
	err = ch.ExchangeDeclare(
		exchangeName, // name
		"fanout",     // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	return &EventBus{
		ch:           make(chan Event, bufferSize),
		rabbitConn:   conn,
		rabbitCh:     ch,
		exchangeName: exchangeName,
	}, nil
}

// Publish 投递事件
func (b *EventBus) Publish(e Event) {
	select {
	case b.ch <- e:
		// 成功放入传送带
	default:
		// 传送带满了，丢弃事件
		log.Printf("[EventBus] ⚠️ Buffer full, dropping event: %s", e.Key)
	}
}

// StartConsumer 启动消费者
func (b *EventBus) StartConsumer() {
	go func() {
		// 监听 Go Channel
		for e := range b.ch {
			// 1. 序列化消息（转成 JSON）
			body, err := json.Marshal(e)
			if err != nil {
				log.Printf("[EventBus] Json Marshal error: %v", err)
				continue
			}

			// 2. 发送到 RabbitMQ
			ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
			err = b.rabbitCh.PublishWithContext(ctx,
			b.exchangeName, 	// exchange
			"",	// routing key
			false,	// mandatory
			false,	// immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body: body,
				Timestamp: time.Now(),
			})
			cancel()

			if err != nil {
				log.Printf("❌ [RabbitMQ] Publish failed: %v", err)
			} else {
				log.Printf("✅ [RabbitMQ] Pushed: %s %s", opStr(e.Type), e.Key)
			}
		}
	}()
}

// Close 优雅关闭
func (b *EventBus) Close() {
	if b.rabbitCh != nil {
		b.rabbitCh.Close()
	}
	if b.rabbitConn != nil {
		b.rabbitConn.Close()
	}
}

func opStr(t EventType) string {
	if t == EventSet {
		return "SET"
	}
	return "DEL"
}
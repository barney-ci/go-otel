package otel

import (
	"github.com/segmentio/kafka-go"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel/propagation"
)

type kafkaCarrier struct {
	hdrs []kafka.Header
}

var _ propagation.TextMapCarrier = (*kafkaCarrier)(nil)

func NewKafkaCarrier(hdrs []kafka.Header) *kafkaCarrier {
	return &kafkaCarrier{
		hdrs: hdrs,
	}
}

func (c *kafkaCarrier) Get(k string) string {
	for _, h := range c.hdrs {
		if h.Key == k {
			return string(h.Value)
		}
	}
	return ""
}

func (c *kafkaCarrier) Set(k, v string) {
	for i, h := range c.hdrs {
		if h.Key == k {
			c.hdrs[i].Value = []byte(v)
			return
		}
	}
	c.hdrs = append(c.hdrs, kafka.Header{Key: k, Value: []byte(v)})
}

func (c *kafkaCarrier) Keys() (keys []string) {
	for _, h := range c.hdrs {
		keys = append(keys, h.Key)
	}
	return
}

func (c *kafkaCarrier) Headers() []kafka.Header {
	return c.hdrs
}

type franzKafkaCarrier struct {
	hdrs []kgo.RecordHeader
}

var _ propagation.TextMapCarrier = (*franzKafkaCarrier)(nil)

func NewFranzKafkaCarrier(hdrs []kgo.RecordHeader) *franzKafkaCarrier {
	return &franzKafkaCarrier{
		hdrs: hdrs,
	}
}

func (c *franzKafkaCarrier) Get(k string) string {
	for _, h := range c.hdrs {
		if h.Key == k {
			return string(h.Value)
		}
	}
	return ""
}

func (c *franzKafkaCarrier) Set(k, v string) {
	for i, h := range c.hdrs {
		if h.Key == k {
			c.hdrs[i].Value = []byte(v)
			return
		}
	}
	c.hdrs = append(c.hdrs, kgo.RecordHeader{Key: k, Value: []byte(v)})
}

func (c *franzKafkaCarrier) Keys() (keys []string) {
	for _, h := range c.hdrs {
		keys = append(keys, h.Key)
	}
	return
}

func (c *franzKafkaCarrier) Headers() []kgo.RecordHeader {
	return c.hdrs
}

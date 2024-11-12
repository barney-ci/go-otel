package otel

import (
	"github.com/segmentio/kafka-go"
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

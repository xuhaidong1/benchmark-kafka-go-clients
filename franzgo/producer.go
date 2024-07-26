package franzgo

import (
	"context"
	"errors"
	"github.com/gguridi/benchmark-kafka-go-clients/config"
	"github.com/gguridi/benchmark-kafka-go-clients/prmths"
	log "github.com/sirupsen/logrus"
	"github.com/twmb/franz-go/pkg/kgo"
	"strconv"
	"sync/atomic"
	"time"
)

var (
	Topic = "topic_adx_test"
)

func NewClient(brokers []string) (*kgo.Client, error) {
	cli, err := kgo.NewClient([]kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.RequiredAcks(kgo.LeaderAck()),
		kgo.ProduceRequestTimeout(time.Minute),
		kgo.DisableIdempotentWrite(),
	}...)

	if err != nil {
		log.WithError(err).Panic("Unable to create franzgo Client")
	}
	return cli, nil
}

func (b *BenchWrapper) Prepare(message []byte, numMessages int) func() {
	log.Infof("Preparing to send message of %d bytes %d times", len(message), numMessages)
	return func() {
		for j := 0; j < numMessages; j++ {
			go func() {
				b.Cli.Produce(context.Background(), &kgo.Record{
					Topic: Topic,
					Value: message,
				}, func(r *kgo.Record, err error) {
					if err != nil {
						log.WithError(err).Panic("Unable to deliver the message")
					} else {
						atomic.AddInt64(&b.SuccessNum, 1)
					}
				})
			}()
		}
		b.Done <- struct{}{}
	}
}

func (b *BenchWrapper) PrepareWithPrometheus(message []byte, numMessages int) func() {
	length := strconv.Itoa(len(message))
	log.Infof("Preparing to send message of %d bytes %d times", len(message), numMessages)
	return func() {
		for j := 0; j < numMessages; j++ {
			go func() {
				start := time.Now().UnixMilli()
				b.Cli.Produce(context.Background(), &kgo.Record{
					Topic: Topic,
					Value: message,
				}, func(r *kgo.Record, err error) {
					if err != nil {
						log.WithError(err).Panic("Unable to deliver the message")
					} else {
						end := time.Now().UnixMilli()
						prmths.ProduceMessageLatency.WithLabelValues("franzgo", length).Observe(float64(end-start) / 1000)
						atomic.AddInt64(&b.SuccessNum, 1)
					}
				})
			}()
		}
		b.Done <- struct{}{}
	}
}

func (b *BenchWrapper) PrepareBench(MessageSize int) func() {
	log.Infof("franzgo Preparing to send message of %d bytes", MessageSize)
	return func() {
		b.Cli.Produce(context.Background(), &kgo.Record{
			Topic: Topic,
			Value: config.GenMessage(MessageSize),
		}, func(r *kgo.Record, err error) {
			if err != nil {
				log.WithError(err).Panic("franzgo Unable to deliver the message")
			} else {
				atomic.AddInt64(&b.SuccessNum, 1)
			}
		})
	}
}

type BenchWrapper struct {
	Cli        *kgo.Client
	SuccessNum int64
	Done       chan struct{}
}

func NewBenchWrapper() *BenchWrapper {
	cli, _ := NewClient(config.Brokers)
	return &BenchWrapper{
		Cli:  cli,
		Done: make(chan struct{}, 1),
	}
}

func (b *BenchWrapper) WaitSignal(ctx context.Context, numMessages int64) {
	var ok bool
	for !ok {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				log.Infof("franzgo plan produce %d messages timeout...real messages %d", numMessages, b.SuccessNum)
			}
			ok = true
		default:
			if atomic.CompareAndSwapInt64(&b.SuccessNum, numMessages, int64(0)) {
				<-b.Done
				b.Cli.Close()
				ok = true
			}
		}
	}
}

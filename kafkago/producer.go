package kafkago

import (
	"context"
	"errors"
	"github.com/gguridi/benchmark-kafka-go-clients/config"
	"github.com/gguridi/benchmark-kafka-go-clients/prmths"
	"strconv"
	"sync/atomic"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	log "github.com/sirupsen/logrus"
)

// NewProducer returns a new kafkago writer.
func NewProducer(brokers []string) *kafkago.Writer {
	//kafkago.Conn{} kafkago.Client
	return &kafkago.Writer{
		Addr:         kafkago.TCP(brokers...),
		Topic:        config.Topic,
		Balancer:     &kafkago.Hash{},
		BatchTimeout: time.Duration(100) * time.Millisecond,
		RequiredAcks: kafkago.RequireOne,
		//BatchSize:     1000,
		// Async doesn't allow us to know if message has been successfully sent to Kafka.
		Async: true,
	}
}

// Prepare returns a function that can be used during the benchmark as it only
// performs the sending of messages.
func (b *BenchWrapper) Prepare(message []byte, numMessages int) func() {
	log.Infof("Preparing to send message of %d bytes %d times", len(message), numMessages)
	return func() {
		for j := 0; j < numMessages; j++ {
			go func() {
				err := b.Cli.WriteMessages(context.Background(), kafkago.Message{Value: message})
				if err != nil {
					log.WithError(err).Panic("Unable to deliver the message")
				} else {
					atomic.AddInt64(&b.SuccessNum, 1)
				}
			}()
		}
		b.Done <- struct{}{}
	}
}

func (b *BenchWrapper) PrepareWithPrometheus(message []byte, numMessages int) func() {
	b.Cli.Async = false
	length := strconv.Itoa(len(message))

	log.Infof("Preparing to send message of %d bytes %d times", len(message), numMessages)
	return func() {
		for j := 0; j < numMessages; j++ {
			go func() {
				start := time.Now().UnixMilli()
				defer func() {
					end := time.Now().UnixMilli()
					prmths.ProduceMessageLatency.WithLabelValues("kafkago", length).Observe(float64(end-start) / 1000)
				}()
				err := b.Cli.WriteMessages(context.Background(), kafkago.Message{Value: message})
				if err != nil {
					log.WithError(err).Panic("Unable to deliver the message")
				} else {
					atomic.AddInt64(&b.SuccessNum, 1)
				}
			}()
		}
		b.Done <- struct{}{}
	}
}

type BenchWrapper struct {
	Done       chan struct{}
	SuccessNum int64
	Cli        *kafkago.Writer
}

func (b *BenchWrapper) PrepareBench(messageSize int) func() {
	b.Cli.Async = false
	log.Infof("Preparing to send message of %d bytes", messageSize)
	return func() {
		go func() {
			err := b.Cli.WriteMessages(context.Background(), kafkago.Message{Value: config.GenMessage(messageSize)})
			if err != nil {
				log.WithError(err).Panic("Unable to deliver the message")
			} else {
				atomic.AddInt64(&b.SuccessNum, 1)
			}
		}()
	}
}

func NewBenchWrapper() *BenchWrapper {
	return &BenchWrapper{
		Done: make(chan struct{}, 1),
		Cli:  NewProducer(config.Brokers),
	}
}

func (b *BenchWrapper) WaitSignal(ctx context.Context, numMessages int64) {
	var ok bool
	for !ok {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				log.Infof("kafkago plan produce %d messages timeout...real messages %d", numMessages, b.SuccessNum)
			}
			ok = true
		default:
			if atomic.CompareAndSwapInt64(&b.SuccessNum, numMessages, int64(0)) {
				<-b.Done
				err := b.Cli.Close()
				if err != nil {
					panic(err)
				}
				ok = true
			}
		}
	}
}

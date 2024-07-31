package sarama

import (
	"context"
	"errors"
	"github.com/gguridi/benchmark-kafka-go-clients/config"
	"github.com/gguridi/benchmark-kafka-go-clients/prmths"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/IBM/sarama"

	log "github.com/sirupsen/logrus"
)

// NewProducer returns a new Sarama async producer.
func NewProducer(cli sarama.Client) sarama.AsyncProducer {
	producer, err := sarama.NewAsyncProducerFromClient(cli)
	if err != nil {
		log.WithError(err).Panic("Unable to start the producer")
	}
	return producer
}

// Prepare returns a function that can be used during the benchmark as it only
// performs the sending of messages, checking that the sending was successful.
func (b *BenchWrapper) Prepare(message []byte, numMessages int) func() {
	log.Infof("Preparing to send message of %d bytes %d times", len(message), numMessages)
	go func() {
		for err := range b.Producer.Errors() {
			log.WithError(err).Panic("Unable to deliver the message")
		}
	}()
	return func() {
		for j := 0; j < numMessages; j++ {
			go func() {
				b.Producer.Input() <- &sarama.ProducerMessage{
					Topic: config.Topic,
					Value: sarama.ByteEncoder(message),
				}
			}()
		}
		b.Done <- struct{}{}
	}
}

func (b *BenchWrapper) PrepareWithPrometheus(message []byte, numMessages int) func() {

	log.Infof("Preparing to send message of %d bytes %d times", len(message), numMessages)

	go func() {
		for err := range b.Producer.Errors() {
			log.WithError(err).Panic("Unable to deliver the message")
		}
	}()

	return func() {
		for j := 0; j < numMessages; j++ {
			go func() {
				b.Producer.Input() <- &sarama.ProducerMessage{
					Topic: config.Topic,
					Metadata: map[string]int64{
						"sendTime": time.Now().UnixMilli(),
						"length":   int64(len(message)),
					},
					Value: sarama.ByteEncoder(message),
				}
			}()
		}
		b.Done <- struct{}{}
	}
}

func NewClient(brokers []string) sarama.Client {
	//log.Infof("Connecting to %s", brokers)
	cli, err := sarama.NewClient(brokers, saramaConfig())
	if err != nil {
		panic(err)
	}
	return cli
}

func (b *BenchWrapper) PrepareBench(messageSize int) func() {
	// log.Infof("sarama Preparing to send message of %d bytes", messageSize)

	go func() {
		for err := range b.Producer.Errors() {
			log.WithError(err).Println("Unable to deliver the message")
		}
	}()

	return func() {
		b.Producer.Input() <- &sarama.ProducerMessage{
			Topic: config.Topic,
			Value: sarama.ByteEncoder(config.GenMessage(messageSize)),
		}
	}
}

type BenchWrapper struct {
	Done       chan struct{}
	Cli        sarama.Client
	Producer   sarama.AsyncProducer
	SuccessNum int64
}

func saramaConfig() *sarama.Config {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V0_10_2_0
	cfg.Producer.RequiredAcks = sarama.WaitForLocal
	cfg.Producer.Return.Successes = true
	cfg.Producer.Flush.Frequency = time.Duration(100) * time.Millisecond
	cfg.Consumer.Return.Errors = true
	cfg.Consumer.Offsets.AutoCommit.Enable = true
	cfg.Consumer.Group.Session.Timeout = time.Minute
	cfg.Consumer.Group.Heartbeat.Interval = 10 * time.Second
	cfg.Net.DialTimeout = time.Minute
	cfg.Net.ReadTimeout = time.Minute
	return cfg
}

func (b *BenchWrapper) WaitSignal(ctx context.Context, numMessages int64) {
	var ok bool
	for !ok {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				log.Infof("sarama plan produce %d messages timeout...real messages %d", numMessages, b.SuccessNum)
			}
			ok = true
		case <-b.Producer.Successes():
			atomic.AddInt64(&b.SuccessNum, 1)
			if atomic.CompareAndSwapInt64(&b.SuccessNum, numMessages, int64(0)) {
				//log.Infof("Sent %d messages... stopping...", numMessages)
				<-b.Done
				if err := b.Producer.Close(); err != nil {
					log.WithError(err).Panic("Unable to close the producer")
				}
				if err := b.Cli.Close(); err != nil {
					log.WithError(err).Panic("Unable to close the client")
				}
				ok = true
			}
		}
	}
}

func (b *BenchWrapper) WaitSignalPrometheus(ctx context.Context, numMessages int64) {
	var ok bool
	for !ok {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				log.Infof("sarama plan produce %d messages timeout...real messages %d", numMessages, b.SuccessNum)
			}
			ok = true
		case msg := <-b.Producer.Successes():
			atomic.AddInt64(&b.SuccessNum, 1)
			metadata := msg.Metadata.(map[string]int64)
			length := strconv.Itoa(int(metadata["length"]))
			sendTime := time.UnixMilli(metadata["sendTime"])
			prmths.ProduceMessageLatency.WithLabelValues("sarama", length).Observe(float64(time.Since(sendTime).Milliseconds() / 1000))

			if atomic.CompareAndSwapInt64(&b.SuccessNum, numMessages, int64(0)) {
				log.Infof("Sent %d messages... stopping...", numMessages)
				<-b.Done
				if err := b.Producer.Close(); err != nil {
					log.WithError(err).Panic("Unable to close the producer")
				}
				if err := b.Cli.Close(); err != nil {
					log.WithError(err).Panic("Unable to close the client")
				}
				ok = true
			}
		}
	}
}

func NewBenchWrapper() *BenchWrapper {
	cli := NewClient(config.Brokers)
	producer := NewProducer(cli)
	return &BenchWrapper{
		Cli:      cli,
		Producer: producer,
		Done:     make(chan struct{}, 1),
	}
}

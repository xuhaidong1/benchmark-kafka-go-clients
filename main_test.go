package main

import (
	"context"
	"fmt"
	"github.com/gguridi/benchmark-kafka-go-clients/config"
	"github.com/gguridi/benchmark-kafka-go-clients/franzgo"
	"github.com/gguridi/benchmark-kafka-go-clients/kafkago"
	"github.com/gguridi/benchmark-kafka-go-clients/sarama"
	. "github.com/onsi/ginkgo"
	"strconv"
	"time"
)

var _ = Describe("Benchmarks", func() {
	Context("NumMessages:1000 MessageSize:100", func() {
		NumMessages = 1000
		MessageSize = 100
		Measure("NumMessages:"+strconv.Itoa(NumMessages)+"MessageSize:"+strconv.Itoa(MessageSize)+"sarama", func(b Benchmarker) {
			Library = "sarama"
			name := fmt.Sprintf("%s producing %d messages of %d bytes size", Library, NumMessages, MessageSize)
			br := sarama.NewBenchWrapper()
			process := br.Prepare(config.GenMessage(MessageSize), NumMessages)
			b.Time(name, func() {
				process()
			})
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
			br.WaitSignal(ctx, int64(NumMessages))
			cancel()
		}, 3)

		Measure("NumMessages:"+strconv.Itoa(NumMessages)+"MessageSize:"+strconv.Itoa(MessageSize)+"kafkago", func(b Benchmarker) {
			Library = "kafkago"
			name := fmt.Sprintf("%s producing %d messages of %d bytes size", Library, NumMessages, MessageSize)
			br := kafkago.NewBenchWrapper()
			process := br.Prepare(config.GenMessage(MessageSize), NumMessages)
			b.Time(name, func() {
				process()
			})
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
			br.WaitSignal(ctx, int64(NumMessages))
			cancel()
		}, 3)

		Measure("NumMessages:"+strconv.Itoa(NumMessages)+"MessageSize:"+strconv.Itoa(MessageSize)+"franzgo", func(b Benchmarker) {
			Library = "franzgo"
			name := fmt.Sprintf("%s producing %d messages of %d bytes size", Library, NumMessages, MessageSize)
			br := franzgo.NewBenchWrapper()
			process := br.Prepare(config.GenMessage(MessageSize), NumMessages)
			b.Time(name, func() {
				process()
			})
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
			br.WaitSignal(ctx, int64(NumMessages))
			cancel()
		}, 3)
	})
})

var MessageSizeNum = map[int][]int{
	1024:  []int{1000, 10000, 100000, 1000000},
	5120:  []int{100, 1000, 10000, 100000, 1000000},
	10240: []int{100, 1000, 10000, 100000}, //max 1G
}

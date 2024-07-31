package main

import (
	"context"
	"fmt"
	"github.com/gguridi/benchmark-kafka-go-clients/franzgo"
	"github.com/gguridi/benchmark-kafka-go-clients/kafkago"
	"github.com/gguridi/benchmark-kafka-go-clients/sarama"
	"sync/atomic"
	"testing"
	"time"
)

var bms = 1000 // 1000 10000 100000

func BenchmarkFranzgo(b *testing.B) {
	br := franzgo.NewBenchWrapper()
	f := br.PrepareBench(bms)
	var sentnum int64
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		atomic.AddInt64(&sentnum, 1)
		f()
	}

	b.StopTimer()
	br.Done <- struct{}{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	br.WaitSignal(ctx, sentnum)
	fmt.Println("BenchmarkFranzgo sentnum", sentnum)
	cancel()
}

func BenchmarkKafkago(b *testing.B) {
	br := kafkago.NewBenchWrapper()
	f := br.PrepareBench(bms)
	var sentnum int64
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		atomic.AddInt64(&sentnum, 1)
		f()
	}

	b.StopTimer()
	br.Done <- struct{}{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	br.WaitSignal(ctx, sentnum)
	fmt.Println("BenchmarkKafkago sentnum", sentnum)
	cancel()
}

func BenchmarkSarama(b *testing.B) {
	br := sarama.NewBenchWrapper()
	f := br.PrepareBench(bms)
	var sentnum int64
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		atomic.AddInt64(&sentnum, 1)
		f()
	}

	b.StopTimer()
	br.Done <- struct{}{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*1)
	br.WaitSignal(ctx, sentnum)
	fmt.Println("BenchmarkSarama sentnum", sentnum)
	cancel()
}

// go test -bench=^BenchmarkSarama$ . -benchmem -cpuprofile=sarama_cpu.pprof
//go test -bench=^BenchmarkKafkago$ -benchtime=60s . -benchmem -cpuprofile=kafkago_cpu.pprof
//go test -bench=^BenchmarkFranzgo$ -benchtime=60s . -benchmem -cpuprofile=franzgo_cpu.pprof

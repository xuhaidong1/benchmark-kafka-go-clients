package main

import (
	"context"
	"fmt"
	"github.com/gguridi/benchmark-kafka-go-clients/config"
	"github.com/gguridi/benchmark-kafka-go-clients/franzgo"
	"github.com/gguridi/benchmark-kafka-go-clients/kafkago"
	"github.com/gguridi/benchmark-kafka-go-clients/sarama"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var (
	// Library contains the library name we are going to run the benchmarks for.
	Library string
	// Brokers contains the list of brokers, comma-separated, to use.
	Brokers = []string{"n-adx-kafka-1.adx.opera.com:1901", "n-adx-kafka-2.adx.opera.com:1902", "n-adx-kafka-3.adx.opera.com:1903"}
	// Topic contains the topic to use in this test.
	Topic = "topic_adx_test"
	//NumMessages contains the number of messages to send.
	NumMessages int
	// MessageSize contains the size of the message to send.
	MessageSize int
)

func init() {
	rand.Seed(time.Now().UnixNano())
	//flag.StringVar(&Library, "library", "", "Library to use for this benchmark")
	//flag.IntVar(&NumMessages, "num", 1000, "Number of messages to send")
	//flag.IntVar(&MessageSize, "size", 1000, "Number of messages to send")
}

func main() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		_ = http.ListenAndServe(":2115", nil)
	}()
	r := gin.Default()
	r.POST("/run", Handler)
	err := r.Run(":8280")
	if err != nil {
		panic(err)
	}
}

func Handler(c *gin.Context) {
	type Req struct {
		Library     string `json:"library"`
		MessageSize int    `json:"message_size"`
		MessageNum  int    `json:"message_num"`
	}
	var req Req
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	switch req.Library {
	case "sarama":
		PrometheusSarama(req.MessageSize, req.MessageNum)
	case "kafkago":
		PrometheusKafkago(req.MessageSize, req.MessageNum)
	case "franzgo":
		PrometheusFranzgo(req.MessageSize, req.MessageNum)
	case "all":
		wg := &sync.WaitGroup{}
		wg.Add(3)
		go func() {
			PrometheusSarama(req.MessageSize, req.MessageNum)
			wg.Done()
		}()
		go func() {
			PrometheusKafkago(req.MessageSize, req.MessageNum)
			wg.Done()
		}()
		go func() {
			PrometheusFranzgo(req.MessageSize, req.MessageNum)
			wg.Done()
		}()
		wg.Wait()
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid library"})
	}
	c.String(http.StatusOK, "Finished")
}

func PrometheusFranzgo(size, num int) {
	br := franzgo.NewBenchWrapper()
	f := br.PrepareWithPrometheus(config.GenMessage(size), num)
	f()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	br.WaitSignal(ctx, int64(num))
	fmt.Println("Franzgo sentnum", num)
	cancel()
}

func PrometheusKafkago(size, num int) {
	br := kafkago.NewBenchWrapper()
	f := br.PrepareWithPrometheus(config.GenMessage(size), num)
	f()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	br.WaitSignal(ctx, int64(num))
	fmt.Println("Kafkago sentnum", num)
	cancel()
}

func PrometheusSarama(size, num int) {
	br := sarama.NewBenchWrapper()
	f := br.PrepareWithPrometheus(config.GenMessage(size), num)
	f()
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	br.WaitSignalPrometheus(ctx, int64(num))
	fmt.Println("BenchmarkSarama sentnum", num)
	cancel()
}

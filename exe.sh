echo "start producer bench test: sarama 1000*100bytes"
ginkgo -focus=producer ./... -- -test.bench=. -library=sarama -num=1000 -size=100
echo "end producer bench test: sarama 1000*100bytes"
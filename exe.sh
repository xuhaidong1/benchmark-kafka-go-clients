#!/bin/zsh
echo "Starting producer test"
#ginkgo ./... -- -test.bench=. -num=1000 -size=1024
#ginkgo ./... -- -test.bench=. -num=10000 -size=1024
#ginkgo ./... -- -test.bench=. -num=100000 -size=1024
ginkgo ./... -- -test.bench=. -num=1000000 -size=1024
ginkgo ./... -- -test.bench=. -num=1000000 -size=2048
#ginkgo ./... -- -test.bench=. -num=1000 -size=5120
#ginkgo ./... -- -test.bench=. -num=10000 -size=5120
#ginkgo ./... -- -test.bench=. -num=100000 -size=5120
#ginkgo ./... -- -test.bench=. -num=1000 -size=10240
#ginkgo ./... -- -test.bench=. -num=10000 -size=10240
#ginkgo ./... -- -test.bench=. -num=100000 -size=10240
echo "Finished producer test"
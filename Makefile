.PHONY: build run run-enhanced clean

build:
	go mod tidy
	go build -o lscc-benchmark main.go

run:
	./lscc-benchmark

run-enhanced:
	go build -o lscc-benchmark main_enhanced.go
	./lscc-benchmark

clean:
	rm -f lscc-benchmark
	rm -f results_*.csv


APP_NAME = lscc-benchmark
CLI_NAME = lscc-cli

build:
	go build -o $(APP_NAME) main.go

cli:
	go build -o $(CLI_NAME) lscc-cli.go

run-nodes:
	chmod +x run-nodes.sh
	./run-nodes.sh

clean:
	rm -f $(APP_NAME) $(CLI_NAME)
	rm -rf logs/*


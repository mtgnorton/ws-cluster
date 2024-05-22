.PHONY: build

NAME = "ws-cluster"

build:
	go build -x -o $(NAME) main.go
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -x -o $(NAME) main.go
kill:
	kill -9 `ps -ef | grep $(NAME) | grep -v grep | awk '{print $$2}'`;
run-dev:
	nohup ./$(NAME) --queue redis  --env dev  &
run-prod:
	nohup ./$(NAME) --queue redis  --env prod  &
ps:
	ps -ef | grep $(NAME)
tail-log:
	tail -n 100 -f logs/normal.log

run-wikitrade:
	go build -o  examples/wikitrade/ws-demo-server examples/wikitrade/server.go;
	nohup ./examples/wikitrade/ws-demo-server  &

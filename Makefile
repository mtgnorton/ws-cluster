.PHONY: build

NAME = "ws-cluster"
build:
	go build -o $(NAME) main.go

run:
	kill -9 `ps -ef | grep $(NAME) | grep -v grep | awk '{print $$2}'`;
	nohup ./$(NAME) --queue redis  --env dev &
ps:
	ps -ef | grep $(NAME)
tail-log:
	tail -n 100 -f logs/normal.log

run-wikitrade:
	go build -o  examples/wikitrade/ws-demo-server examples/wikitrade/server.go;
	nohup ./examples/wikitrade/ws-demo-server  &

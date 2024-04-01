.PHONY: build

NAME = "ws-cluster"
build:
	go build -o $(NAME) main.go

run:
	nohup ./$(NAME) --queue redis  --env dev &

tail-log:
	tail -n 100 -f logs/normal.log

run-wikitrade:
	go build -o  examples/wikitrade/ws-demo-server examples/wikitrade/server.go;
	nohup ./examples/wikitrade/ws-demo-server  &

.PHONY: build

NAME = "ws-cluster"

build:
	go build -x -o bin/$(NAME) main.go
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -x -o bin/$(NAME)-linux main.go
kill:
	kill -9 `ps -ef | grep $(NAME) | grep -v grep | awk '{print $$2}'`;
run-local:build
	./bin/$(NAME) --queue redis  --env local 
run-dev:
	nohup ./$(NAME) --queue redis  --env dev  &
run-prod:
	nohup ./$(NAME) --queue redis  --config conf/config.prod.yaml --env prod  &
ps:
	ps -ef | grep $(NAME)
tail-log:
	tail -n 100 -f logs/normal.log

run-wikitrade:
	go build -o  examples/wikitrade/ws-demo-server examples/wikitrade/server.go;
	nohup ./examples/wikitrade/ws-demo-server  &
build-docker:build-linux
	docker build \
	--platform linux/amd64 \
	--build-arg "HTTP_PROXY=http://host.docker.internal:7890/" \
	--build-arg "HTTPS_PROXY=http://host.docker.internal:7890/" \
	--build-arg "APP_NAME=ws-cluster" \
	--build-arg "MAIN_PATH=main.go" \
	--build-arg "CONFIG_PATH=conf/config.docker.official.yaml" \
	--build-arg "CONFIG_FILE_NAME=config.docker.official.yaml" \
	-t mtgnorton/ws-cluster:latest -f Dockerfile .

run-docker:
	docker run --rm --name ws-cluster -p 8084:8084 mtgnorton/ws-cluster:latest --queue redis --config config.docker.yaml

push-docker:
	docker push mtgnorton/ws-cluster:latest

bp-docker:build-docker push-docker

restart-k8s:
	kubectl rollout restart deployment/ws-cluster-deployment

bp-restart:bp-docker restart-k8s

scp:build-linux
	scp bin/ws-cluster-linux  trade-official:/home/ws-cluster/ws-cluster
scp-config:
	scp conf/config-linux.yaml trade-official:/home/ws-cluster/
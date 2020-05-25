# make helper commands
check_golang_exist:
	@[ `command -v go` ] || (echo "Golang is required." && exit 1)

check_docker_exist:
	@[ `command -v docker` ] || (echo "Docker is required." && exit 1)

check_docker_compose_exist:
	@[ `command -v docker-compose` ] || (echo "Docker-Compose is required." && exit 1)

# Docker commands
generate_protobuf: check_docker_exist
	docker run --rm -v `pwd`:/proto -w `pwd` znly/protoc -I /proto/protobuf plates.proto --go_out=plugins=grpc:/proto/client/internal/proto --go_out=plugins=grpc:/proto/server/internal/proto

docker_client: check_docker_exist
	cd client; docker run --rm --name app_client -d --network=host -v `pwd`:/project -w /project golang bash -c 'go get -d ./...; go run main.go'

docker_server: check_docker_exist
	cd server; docker run --rm --name app_server -d --network=host -e DB_HOST=localhost -e DB_PASSWORD=secretpass -v `pwd`:/project -w /project golang bash -c 'go get -d ./...; go run main.go'

docker_db:
	docker run --rm --name app_db -d -p 5432:5432 -e POSTGRES_PASSWORD=secretpass postgres:alpine

# gRPC Client/Server commands
run_client: check_golang_exist
	cd client && \
	go get -d ./... && \
	go run main.go

run_server_with_stdout: check_golang_exist
	cd server && \
	export DB_PASSWORD=secretpass && \
	go get -d ./... && \
	go run main.go

run_server_with_db: check_golang_exist
	cd server && \
	export OUTPUT=db && \
	export DB_PASSWORD=secretpass && \
	go get -d ./... && \
	go run main.go

# Docker-Compose commands
stack_up: check_docker_compose_exist
	docker-compose up --build -d

stack_down: check_docker_compose_exist
	docker-compose down

stack_logs: check_docker_compose_exist
	docker-compose logs -f
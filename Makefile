build:
	go build -o ky

gen-proto:
	protoc --go_out=. --go_opt=paths=source_relative \
    	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
    	internal/plugins/plugins.proto

run-plugins-server: build
	./ky plugins server --bucket="krayon-plugins"
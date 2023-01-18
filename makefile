.Phony: gen protocopy clean tests server client
serverport=8080
gen: 
	protoc -I proto --go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	proto/*.proto

protocopy:
	cp proto/*.proto ../pcbook-Java/app/src/main/proto

clean:
	rm pb/*.go

tests:
	go test -cover -race -timeout 1s ./...

server:
	go run cmd/server/*.go -serverport ${serverport}

client:
	go run cmd/client/*.go -serverport ${serverport}
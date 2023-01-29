.PHONY: gen protocopy clean tests server client evan cert serverTLSREST clientTLS serverGRPC serverTLSGRPC clientTLS
serverport1=50051
serverrest=50052
serverport=8080
nginx=8080
gen:  
	protoc --proto_path=proto proto/*.proto  --go_out=:pb --go-grpc_out=:pb --grpc-gateway_out=:pb --openapiv2_out=:swagger

protocopy:
	@echo "If error need fix path"
	cp proto/*.proto ../pcbook-Java/src/main/proto

clean:
	rm pb/*.go

tests:
	go test -cover -race -timeout 1s ./...

serverGRPC:
	go run cmd/server/*.go -serverport ${serverport1}

serverREST:
	go run cmd/server/*.go -serverport ${serverrest} -rest

serverTLSREST:
	go run cmd/server/*.go -serverport ${serverrest} -tls -rest

serverTLSGRPC:
	go run cmd/server/*.go -serverport ${serverport1} -tls

client:
	go run cmd/client/*.go -serverport ${nginx}

clientTLS:
	go run cmd/client/*.go -serverport ${nginx} -tls

evan:
	@echo "to install please visit https://github.com/ktr0731/evans"
	evans -r repl -p ${serverport}

cert:
	cd cert; chmod +x gen.sh; ./gen.sh; cd ..
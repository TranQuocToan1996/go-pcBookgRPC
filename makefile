.PHONY: gen protocopy clean tests server client evan cert serverTLS clientTLS server1
serverport1=50051
serverport2=50052
serverport=8080
nginx=8080
gen:  
	protoc -I proto --go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	proto/*.proto

protocopy:
	@echo "If error need fix path"
	cp proto/*.proto ../pcbook-Java/src/main/proto

clean:
	rm pb/*.go

tests:
	go test -cover -race -timeout 1s ./...

server:
	go run cmd/server/*.go -serverport ${serverport1}

server1:
	go run cmd/server/*.go -serverport ${serverport2}

serverTLS:
	go run cmd/server/*.go -serverport ${serverport1} -tls

serverTLS1:
	go run cmd/server/*.go -serverport ${serverport2} -tls

client:
	go run cmd/client/*.go -serverport ${nginx}

clientTLS:
	go run cmd/client/*.go -serverport ${nginx} -tls

evan:
	@echo "to install please visit https://github.com/ktr0731/evans"
	evans -r repl -p ${serverport}

cert:
	cd cert; chmod +x gen.sh; ./gen.sh; cd ..
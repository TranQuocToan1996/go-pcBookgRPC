# pcBookgRPC Golang
- Special thanks to: [TECH SCHOOL](https://www.youtube.com/@TECHSCHOOLGURU)
- Summary: Server/client to manage and search laptop configurations. It provides 4 gRPC/REST APIs included unit tests:

1. Create a new laptop: unary gRPC
    Allows client to create a new laptop with some specific configurations.

2. Search laptops with some filtering conditions: server-streaming gRPC
    Allows client to search for laptops that satisfies some filtering conditions.

3. Upload a laptop image file in chunks: client-streaming gRPC
   Allows client to upload 1 laptop image file to the server. The file will be split into multiple chunks, and they will be sent to the server as a stream.

4. Rate multiple laptops and get back average rating for each of them: bidirectional-streaming gRPC
    Allows client to rate multiple laptops, each with a score, and get back the average rating score for each of them.

- [Java version](https://github.com/TranQuocToan1996/pcbook-Java) (Server/client): https://github.com/TranQuocToan1996/pcbook-Java

- First running:
    1. Setup SSL/TLS. 
        + Generates certificate/key in cert folder.
        + Config tls.Config struct as part of grpc options (client and server).
        + Config nginx.conf
        + Config ports/flags in makefile (client/server)

    2. Install dependencies: 
        + At root dir.
        ```
        go get ./... && go mod tidy
        ```
        + Install 3rd dependencies gRPC APIs, nginx,...

    3. Start REST/gRPC server: -rest flag for REST one. Default gRPC one.

    4. Client calling:
        + gRPC: Use [evans](https://github.com/ktr0731/evans) or clients in Go/Java to call.
        + REST: curl, [REST client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) or [Postman](https://www.postman.com/).

- TODO tasks:
    1. Connect to Database.
    2. Move const variable to config file/env var.
    3. Change: https://github.com/dgrijalva/jwt-go -> https://github.com/golang-jwt/jwt due to security problem.
    
    jwt-go allows attackers to bypass intended access restrictions in situations with []string{} for m["aud"] (which is allowed by the specification). Because the type assertion fails, "" is the value of aud. This is a security problem if the JWT token is presented to a service that lacks its own audience check. There is no patch available and users of jwt-go are advised to migrate to golang-jwt at version 3.2.1
    
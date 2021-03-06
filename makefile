install:
	@echo building and installing inscript
	go install
	@echo "installed inscript under ${GOBIN} (\$$GOBIN) or ${GOPATH}/bin (\$$GOPATH/bin)"

all: test install

fmt:
	@go fmt
	@go fmt ./lexer
	@go fmt ./parser
	@go fmt ./ast
	@go fmt ./token
	@go fmt ./runtime

test:
	go test ./lexer
	go test ./parser

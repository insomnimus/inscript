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

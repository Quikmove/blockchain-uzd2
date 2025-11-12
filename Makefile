build:
	@go build -o bin/cli ./cmd/cli | true



clean:
	@rm -rf bin/ | true

BINARY=golangApi
.DEFAULT_GOAL := run
dev:
	rm -rf my.db
	redis-cli flushall
	go build -ldflags "-X main.devenv=development"
	./${BINARY}
BINARY=golangApi
.DEFAULT_GOAL := run
dev:
	rm -rf my.db
	go build -ldflags "-X main.devenv=development"
	./${BINARY}

# Go parameters
GOCMD=go
GORUN=$(GOCMD) run
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
MAIN="./cmd/gohltb/main.go"
BINARY_NAME="bin/gohltb"

build:
	$(GOBUILD) -o $(BINARY_NAME) -v $(MAIN)

test:
	$(GOTEST) -v ./

run:
	$(GORUN) cmd/gohltb/main.go

compile32:
	# 32-Bit Systems
	# FreeBDS
	GOOS=freebsd GOARCH=386 $(GOBUILD) -o ${BINARY_NAME}-freebsd-386 $(MAIN)
	# Linux
	GOOS=linux GOARCH=386 $(GOBUILD) -o ${BINARY_NAME}-linux-386 $(MAIN)
	# Windows
	GOOS=windows GOARCH=386 $(GOBUILD) -o ${BINARY_NAME}-windows-386 $(MAIN)

compile64:
	# 64-Bit
	# FreeBDS
	GOOS=freebsd GOARCH=amd64 $(GOBUILD) -o ${BINARY_NAME}-freebsd-amd64 $(MAIN)
	# MacOS
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o ${BINARY_NAME}-darwin-amd64 $(MAIN)
	# Linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o ${BINARY_NAME}-linux-amd64 $(MAIN)
	# Windows
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o ${BINARY_NAME}-windows-amd64 $(MAIN)

compile_all:
	compile32 compile64
VERSION := 1.0.1
BINARY := osir
LDFLAGS := -s -w -X main.version=$(VERSION)
DIST := dist

.PHONY: build build-all install test clean

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

build-all: clean
	@mkdir -p $(DIST)
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)-linux-amd64 .
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)-linux-arm64 .
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)-darwin-amd64 .
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(DIST)/$(BINARY)-windows-amd64.exe .
	@echo "Built binaries in $(DIST)/"
	@ls -lh $(DIST)/

install:
	go install -ldflags "$(LDFLAGS)" .

test:
	go test ./...

clean:
	rm -rf $(DIST) $(BINARY) $(BINARY).exe

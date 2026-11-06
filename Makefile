.PHONY: all build test clean run-daemon run-ui

all: test build

build:
	go build -o bin/quickcast ./cmd/quickcast
	go build -o bin/quickcastd ./cmd/quickcastd

test:
	go test -v ./...

clean:
	rm -rf bin/

run-daemon: build
	./bin/quickcastd

run-ui:
	quickshell -p ./ui

install: build
	install -Dm755 bin/quickcast $(DESTDIR)/usr/local/bin/quickcast
	install -Dm755 bin/quickcastd $(DESTDIR)/usr/local/bin/quickcastd

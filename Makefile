OUTPUT=userd

SOURCE := $(shell find . -name '*.go')
GOPATH := $(shell pwd)

.PHONY=all clean run-tests

all: $(OUTPUT)

clean:
	rm -rf $(OUTPUT) src

$(OUTPUT): src $(SOURCE)
	GOPATH=$(GOPATH) go build .

src:
	GOPATH=$(GOPATH) go get -d .


run-tests: $(OUTPUT)
	./bin/run-tests.sh
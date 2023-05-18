export GOBIN=$(PWD)/bin

.PHONY: build clean default
default: build
build:
	@mkdir -p bin/
	@go install -v ./...
clean:
	@rm -rf bin/
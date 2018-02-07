FORK?=github.com/IzakMarais
TARGET:=$(GOPATH)/bin/grafana-reporter
SRC:=$(GOPATH)/src/$(FORK)/reporter

.PHONY: build
build: $(TARGET)

.PHONY: docker-build
docker-build:
	@docker build -t grafana-reporter:latest .

.PHONY: test
test:
	@go test -v ./...

$(GOPATH)/bin/dep:
	@go get -u github.com/golang/dep/cmd/dep

$(TARGET): $(GOPATH)/bin/dep
	@cd $(SRC) && dep ensure
	@cd $(SRC)/cmd/grafana-reporter && go install


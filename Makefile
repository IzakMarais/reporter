TARGET:=$(GOPATH)/bin/grafana-reporter
SRC:=$(GOPATH)/src/github.com/izakmarais/reporter

.PHONY: build
build: $(TARGET)

.PHONY: docker-build
docker-build:
	@docker build -t grafana-reporter:0.6.2 .

.PHONY: test
test:
	@go test -v ./...

$(GOPATH)/bin/dep:
	@go get -u github.com/golang/dep/cmd/dep

$(TARGET): $(GOPATH)/bin/dep
	@cd $(SRC) && dep ensure
	@cd $(SRC)/cmd/grafana-reporter && go install

.PHONY: compose-up
compose-up:
	@docker-compose -f ./util/docker-compose.yml up 

.PHONY: compose-down
compose-down:
	@docker-compose -f ./util/docker-compose.yml stop 
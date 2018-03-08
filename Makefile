TARGET:=$(GOPATH)/bin/grafana-reporter
SRC:=$(GOPATH)/src/github.com/IzakMarais/reporter

.PHONY: build
build: $(TARGET)

.PHONY: docker-build
docker-build:
	@docker build -t izakmarais/grafana-reporter:2.0.0 -t izakmarais/grafana-reporter:latest .

docker-push:
	@docker push izakmarais/grafana-reporter

.PHONY: test
test: $(TARGET)
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

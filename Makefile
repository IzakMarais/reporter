# Running on Windows
#
# Set GOPATH in terminal. Example (make for windows needs forwardslashes):
#   set GOPATH=T:/repos/reporter

TARGET:=$(GOPATH)/bin/grafana-reporter
ifeq ($(OS),Windows_NT)
	TARGET:=$(GOPATH)/bin/grafana-reporter.exe
endif
SRC:=$(GOPATH)/src/github.com/IzakMarais/reporter

.PHONY: buildall
buildall: build buildlinux

.PHONY: build
build: 
	go install -v github.com/IzakMarais/reporter/cmd/grafana-reporter

.PHONY: buildlinux 
buildlinux: 	
	cmd //v //c "set GOOS=linux&&go install -v github.com/IzakMarais/reporter/cmd/grafana-reporter"

.PHONY: clean
clean: 	
	rm -rf $(GOPATH)/bin

.PHONY: docker-build
docker-build:
	@docker build -t izakmarais/grafana-reporter:2.3.0 -t izakmarais/grafana-reporter:latest .

docker-push:
	@docker push izakmarais/grafana-reporter

.PHONY: test
test: $(TARGET)
	@go test -v ./...

.PHONY test2:
	@echo hello $(TARGET)

$(GOPATH)/bin/dep:
	@go get -u github.com/golang/dep/cmd/dep

update-deps: $(GOPATH)/bin/dep
	@cd $(SRC) && dep ensure
	@cd $(SRC)/cmd/grafana-reporter && go install

$(TARGET):
	@cd $(SRC)/cmd/grafana-reporter && go install

.PHONY: compose-up
compose-up:
	@docker-compose -f ./util/docker-compose.yml up

.PHONY: compose-down
compose-down:
	@docker-compose -f ./util/docker-compose.yml stop

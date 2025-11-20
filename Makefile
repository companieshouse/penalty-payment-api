CHS_ENV_HOME ?= $(HOME)/.chs_env
GOPATH ?= $(OLDPWD)
TESTS        ?= ./...
COVERAGE_OUT = coverage.out

bin          := penalty-payment-api
version      ?= unversioned
xunit_output := test.xml
lint_output  := lint.txt
govulncheck   := golang.org/x/vuln/cmd/govulncheck@latest

.EXPORT_ALL_VARIABLES:
GO111MODULE = on

.PHONY:
arch:
	@echo OS: $(shell uname) ARCH: $(shell uname -p)

.PHONY: all
all: build

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: build
build: arch fmt depvulncheck
ifeq ($(shell uname; uname -p), Darwin arm)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ go build --ldflags '-linkmode external -extldflags "-static"' -o ecs-image-build/app/$(bin)
else
	CGO_ENABLED=0 go build -o ecs-image-build/app/$(bin)
endif

.PHONY: test
test: test-unit test-integration

.PHONY: test-unit
test-unit:
	go env -w GOBIN=/usr/local/bin
	@go install github.com/quantumcycle/go-ignore-cov@latest
	@go test -run 'Unit' -coverpkg=./... -coverprofile=$(COVERAGE_OUT) $(TESTS) -json > report.json
	@go-ignore-cov --file $(COVERAGE_OUT)
	@go tool cover -func $(COVERAGE_OUT)

.PHONY: test-integration
test-integration:
	go test $(TESTS) -run 'Integration'

.PHONY: test-with-coverage
test-with-coverage:
	@go get github.com/hexira/go-ignore-cov
	@go build -o ${GOBIN} github.com/hexira/go-ignore-cov
	@go test -coverpkg=./... -coverprofile=$(COVERAGE_OUT) $(TESTS)
	@go-ignore-cov --file $(COVERAGE_OUT)
	@go tool cover -func $(COVERAGE_OUT)
	@make coverage-html

.PHONY: clean-coverage
clean-coverage:
	@rm -f $(COVERAGE_OUT) coverage.html

.PHONY: coverage-html
coverage-html:
	@go tool cover -html=$(COVERAGE_OUT) -o coverage.html

.PHONY: clean
clean: clean-coverage
	go mod tidy
	rm -rf ./ecs-image-build/app ./$(bin)-*.zip $(test_path) build.log

.PHONY: package
package:
ifndef version
	$(error No version given. Aborting)
endif
	$(info Packaging version: $(version))
	$(eval tmpdir := $(shell mktemp -d build-XXXXXXXXXX))
	cp ./ecs-image-build/app/$(bin) $(tmpdir)
	cd $(tmpdir) && zip -r ../$(bin)-$(version).zip $(bin)
	rm -rf $(tmpdir)

.PHONY: dist
dist: clean build package

.PHONY: xunit-tests
xunit-tests: GO111MODULE = off
xunit-tests:
	go get github.com/tebeka/go2xunit
	@set -a; go test -v $(TESTS) -run 'Unit' | go2xunit -output $(xunit_output)

.PHONY: lint
lint: GO111MODULE = off
lint:
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install
	gometalinter ./... > $(lint_output); true

.PHONY: depvulncheck
depvulncheck:
	go install $(govulncheck)
	CGO_ENABLED=1 $(GOPATH)/bin/govulncheck -show verbose ./...

.PHONY: docker-image
docker-image: dist
	chmod +x build-docker-local.sh
	./build-docker-local.sh

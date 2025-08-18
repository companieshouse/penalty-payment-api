CHS_ENV_HOME ?= $(HOME)/.chs_env
TESTS        ?= ./...
COVERAGE_OUT = coverage.out

bin          := penalty-payment-api
xunit_output := test.xml
lint_output  := lint.txt

.EXPORT_ALL_VARIABLES:
GO111MODULE = on

.PHONY: all
all: build

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: build
build: fmt $(bin)

$(bin):
	CGO_ENABLED=0 go build -o ./$(bin)

.PHONY: test
test: test-unit test-integration

.PHONY: test-unit
test-unit:
	go env -w GOBIN=/usr/local/bin
	go install -v -x -a github.com/quantumcycle/go-ignore-cov@latest
	go test -run 'Unit' -coverpkg=./... -coverprofile=$(COVERAGE_OUT) $(TESTS)
	ls -l /usr/local/bin/
	echo $(PATH)

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
	rm -f ./$(bin) ./$(bin)-*.zip $(test_path) build.log

.PHONY: package
package:
ifndef version
	$(error No version given. Aborting)
endif
	$(info Packaging version: $(version))
	$(eval tmpdir := $(shell mktemp -d build-XXXXXXXXXX))
	cp ./$(bin) $(tmpdir)
	cp -r ./assets  $(tmpdir)/assets
	cd $(tmpdir) && zip -r ../$(bin)-$(version).zip $(bin) assets
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

.PHONY: security-check
security-check dependency-check:
	@go get golang.org/x/vuln/cmd/govulncheck
	@go get github.com/sonatype-nexus-community/nancy@latest
	@go list -json -deps ./... | nancy sleuth -o json | jq
	@go build -o ${GOBIN} golang.org/x/vuln/cmd/govulncheck
	@govulncheck ./...

.PHONY: security-check-summary
security-check-summary:
	@go get golang.org/x/vuln/cmd/govulncheck
	@go get github.com/sonatype-nexus-community/nancy@latest
	@LOW=0 MED=0 HIGH=0 CRIT=0 res=`go list -json -deps ./... | nancy sleuth -o json | jq -c '.vulnerable[].Vulnerabilities[].CvssScore'`; for score in $$res; do if [ $${score:1:1} -ge 9 ]; then CRIT=$$(($$CRIT+1)); elif [ $${score:1:1} -ge 7 ]; then HIGH=$$(($$HIGH+1)); elif [ $${score:1:1} -ge 4 ]; then MED=$$(($$MED+1)); else LOW=$$(($$LOW+1)); fi; done; echo -e "CRITICAL=$$CRIT\nHigh=$$HIGH\nMedium=$$MED\nLow=$$LOW";
	@go build -o ${GOBIN} golang.org/x/vuln/cmd/govulncheck
	@OTHER=`govulncheck ./... | grep "More info:" | wc -l | tr -d ' '`; echo -e "\nOther=$$OTHER"

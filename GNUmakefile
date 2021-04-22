SWEEP?=us-central1
TEST?=$$(go list ./...)

default: build

build: fmtcheck generate
	go install

fmt:
	@echo "==> Fixing source code with gofmt..."
	gofmt -w -s ./internal/provider

# Currently required by tf-deploy compile
fmtcheck:
	@echo "==> Checking source code against gofmt..."
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

generate:
	go generate  ./...

lint:
	@echo "==> Checking source code against linters..."
	@golangci-lint run ./internal/provider

sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	go test ./internal/provider -v -sweep=$(SWEEP) -sweep-run=$(SWEEPARGS) -timeout 60m

test: fmtcheck generate
	go test $(TESTARGS) -timeout=30s $(TEST)

# Run acceptance tests
.PHONY: testacc
testacc: fmtcheck
	TF_ACC=1 go test -count=1 $(TEST) -v $(TESTARGS) -timeout 120m
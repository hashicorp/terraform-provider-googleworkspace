TEST?=$$(go list ./...)

default: testacc

# Currently required by tf-deploy compile
fmtcheck:
	@echo "==> Checking source code against gofmt..."
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

generate:
	go generate  ./...

test: fmtcheck generate
	go test $(TESTARGS) -timeout=30s $(TEST)

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m
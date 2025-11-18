
args=
path=./...

GOBIN=$(shell go env GOPATH)/bin

lint: setup
	@$(GOBIN)/staticcheck $(path) $(args)
	@go vet $(BUILD_TAGS) $(path) $(args)
	@$(GOBIN)/errcheck ./...
	@echo "StaticCheck & Go Vet & ErrCheck found no problems on your code!"

test: setup
	$(GOBIN)/richgo test $(path) $(args)

setup: $(GOBIN)/richgo $(GOBIN)/staticcheck $(GOBIN)/errcheck

$(GOBIN)/richgo:
	go install github.com/kyoh86/richgo@latest

$(GOBIN)/staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@latest

$(GOBIN)/errcheck:
	go install github.com/kisielk/errcheck@latest

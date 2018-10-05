default:

clean: 
	@rm -rf dist

ensure-version:
ifndef VERSION
	$(error VERSION is undefined)
endif

release: clean release-linux release-darwin release-windows

release-linux: ensure-version
	@mkdir -p dist/linux
	@GOOS=linux GOARCH=amd64 go build -o dist/linux/genesis ./cmd/genesis
	@tar -czf dist/genesis-$(VERSION)-linux-amd64.tar.gz -C dist/linux genesis
	@rm -rf dist/linux

release-darwin: ensure-version
	@mkdir -p dist/darwin
	@GOOS=darwin GOARCH=amd64 go build -o dist/darwin/genesis ./cmd/genesis
	@tar -czf dist/genesis-$(VERSION)-darwin-amd64.tar.gz -C dist/darwin genesis
	@rm -rf dist/darwin

release-windows: ensure-version
	@mkdir -p dist/windows
	@GOOS=windows GOARCH=amd64 go build -o dist/windows/genesis ./cmd/genesis
	@tar -czf dist/genesis-$(VERSION)-windows-amd64.tar.gz -C dist/windows genesis
	@rm -rf dist/windows

.PHONY: release release-linux release-darwin release-windows

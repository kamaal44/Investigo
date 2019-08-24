# requires go get -u golang.org/x/tools/...

test: imports
	@(go list ./... | grep -v "vendor/" | xargs -n1 go test -v -cover)

imports:
	@(goimports -w Investigo)

fmt:
	@(gofmt -w Investigo)

build: build-linux build-darwin build-windows

build-linux: imports
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o dist/darwin-amd64/investigo -v github.com/lucmski/Investigo

build-darwin: imports
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o dist/linux-amd64/investigo -v github.com/lucmski/Investigo

build-windows: imports
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o dist/windows-amd64/investigo.exe -v github.com/lucmski/Investigo



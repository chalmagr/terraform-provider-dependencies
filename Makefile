TEST?=$$(go list ./... | grep -v 'vendor')
NAMESPACE=chalmagr
NAME=dependencies
VERSION=$(shell cat version.txt)
OS=$(shell go env GOHOSTOS)
ARCH=$(shell go env GOHOSTARCH)
OS_ARCH=${OS}_${ARCH}
BINARY=terraform-provider-${NAME}

default: install

build:
	GOOS=${OS} GOARCH=${ARCH} go build -o ./bin/${OS}_${ARCH}/${BINARY}_v${VERSION}_x5

release:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/darwin_amd64/${BINARY}_v${VERSION}_x5
	GOOS=freebsd GOARCH=386 go build -o ./bin/freebsd_386/${BINARY}_v${VERSION}_x5
	GOOS=freebsd GOARCH=amd64 go build -o ./bin/freebsd_amd64/${BINARY}_v${VERSION}_x5
	GOOS=freebsd GOARCH=arm go build -o ./bin/freebsd_arm/${BINARY}_v${VERSION}_x5
	GOOS=linux GOARCH=386 go build -o ./bin/linux_386/${BINARY}_v${VERSION}_x5
	GOOS=linux GOARCH=amd64 go build -o ./bin/linux_amd64/${BINARY}_v${VERSION}_x5
	GOOS=linux GOARCH=arm go build -o ./bin/linux_arm/${BINARY}_v${VERSION}_x5
	GOOS=openbsd GOARCH=386 go build -o ./bin/openbsd_386/${BINARY}_v${VERSION}_x5
	GOOS=openbsd GOARCH=amd64 go build -o ./bin/openbsd_amd64/${BINARY}_v${VERSION}_x5
	GOOS=solaris GOARCH=amd64 go build -o ./bin/solaris_amd64/${BINARY}_v${VERSION}_x5
	GOOS=windows GOARCH=386 go build -o ./bin/windows_386/${BINARY}_v${VERSION}_x5
	GOOS=windows GOARCH=amd64 go build -o ./bin/windows_amd64/${BINARY}_v${VERSION}_x5


install: build
	mkdir -p ~/.terraform.d/plugins/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${OS_ARCH}/${BINARY}_v${VERSION}_x5

test: 
	go test -i $(TEST) || exit 1                                                   
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4                    

testacc: 
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m   

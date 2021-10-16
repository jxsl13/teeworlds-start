

# build and start the server fleet
default: build

# start the server manager
start: build
	./teeworlds-start

# build alias
build:
	go build -race .

# build for llinux
linux:
	env GOOS=linux GOARCH=amd64 go build -race .

windows:
	env GOOS=windows GOARCH=amd64 go build -race .

macos:
	env GOOS=darwin GOARCH=amd64 go build -race .

macos-arm:
	env GOOS=darwin GOARCH=arm64 go build -race .




# build and start the server fleet
default: build

# start the server manager
start: build
	./teeworlds-start

# build alias
build: teeworlds-start

# build the main.go file
teeworlds-start: main.go
	go build .
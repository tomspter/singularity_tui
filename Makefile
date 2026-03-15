APP := singularity-tui

.PHONY: build build-linux-amd64 clean

build: build-linux-amd64

build-linux-amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(APP) .

clean:
	rm -f $(APP)

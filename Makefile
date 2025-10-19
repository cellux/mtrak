mtrak: $(wildcard *.go)
	go build

.PHONY: clean
clean:
	rm -f mtrak

all:
	go build -o livechan.bin *.go
clean:
	rm -f livechan.bin
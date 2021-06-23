build:
	go build -o ./dist/fire.exe ./cmd/fire
	go build -o ./dist/relay.exe ./cmd/relay

clean:
	rd /s /q dist

clean_linux:
	rm -rf dist

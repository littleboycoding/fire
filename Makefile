build:
	GOOS=linux GOARCH=amd64 go build -o ./dist/fire_linux ./cmd/fire
	GOOS=linux GOARCH=amd64 go build -o ./dist/relay_linux ./cmd/relay

	GOOS=darwin GOARCH=amd64 go build -o ./dist/fire_darwin ./cmd/fire
	GOOS=darwin GOARCH=amd64 go build -o ./dist/relay_darwin ./cmd/relay

	GOOS=windows GOARCH=amd64 go build -o ./dist/fire_window.exe ./cmd/fire
	GOOS=windows GOARCH=amd64 go build -o ./dist/relay_window.exe ./cmd/relay

clean:
	rm -rf ./dist
build:
	go build -o ./dist/fire.exe ./cmd/fire

clean:
	rd /s /q dist
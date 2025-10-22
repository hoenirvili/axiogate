build-linux: 
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags "-w -s" -o axiogate ./cmd/axiogate

build: 
	CGO_ENABLED=0 go build -ldflags "-w -s" -o axiogate ./cmd/axiogate

clean:
	rm axiogate

migrate-up:
	migrate -source file://migrations -database "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" up

migrate-down:
	migrate -source file://migrations -database "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" down -all

docker: build-linux
	docker build -t axiogate:0.1 .

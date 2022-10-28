all:
	go fmt
	go build

test:
	go test -v -coverprofile cover.out
	go tool cover -func=cover.out
	rm cover.out

run:
	go fmt
	go run github.com/ema/pets

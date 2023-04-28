all:
	go fmt
	CGO_ENABLED=0 go build
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o pets-arm
	asciidoctor -b manpage manpage.adoc

test:
	go test -v -coverprofile cover.out
	go tool cover -func=cover.out

cover: test
	go tool cover -html cover.out -o cover.html
	open cover.html &

run:
	go fmt
	go run github.com/ema/pets

clean:
	-rm pets pets.1 cover.out cover.html

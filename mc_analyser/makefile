run:
	pnpm run build && templ generate && go build -o ./bin/main .
build:
	pnpm run build && templ generate && GOOS=linux GOARCH=amd64 go build -o mc_analyser -ldflags="-X 'main.isProduction=true'" .
test:
	go test -v ./...
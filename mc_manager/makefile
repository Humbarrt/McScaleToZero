run:
	pnpm run build && templ generate && go build -o ./bin/main .
build:
	pnpm run build && templ generate && GOOS=linux GOARCH=amd64 go build -o mc_manager -ldflags="-X 'main.version=1.0'" .
test:
	go test -v ./... -shuffle on
dockerize:
	docker buildx build --platform linux/amd64 -t humbarrt/mc_manager:latest .

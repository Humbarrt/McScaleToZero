run:
	go run .
build:
	GOOS=linux GOARCH=amd64 go build -o mc_proxy -ldflags="-X 'main.isProduction=true'" .
deploy:
	rsync -avz -e "ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" --progress ./mc_proxy root@91.107.192.240:/root/mc_proxy/
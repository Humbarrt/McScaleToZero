run:
	go run main.go
build:
	GOOS=linux GOARCH=amd64 go build -o mc_server -ldflags="-X 'main.Version=prod'" .
deploy:
	rsync -avz -e "ssh -i ./data/mc-server-key.pem -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" --progress ./mc_server ec2-user@$(shell make get_ip):/home/ec2-user/mc_server/
connect:
	ssh -i "./data/mc-server-key.pem" -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no ec2-user@$(shell make get_ip)
get_ip:
	@aws ec2 describe-instances --filters "Name=tag:Project,Values=minecraft" --query 'Reservations[*].Instances[0].PublicIpAddress' --output text | tr -d '\n'

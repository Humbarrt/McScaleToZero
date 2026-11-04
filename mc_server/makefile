connect:
	sudo ssh -i "./data/aws-server-key.pem" -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no ec2-user@$(shell make get_ip)
get_ip:
	@aws ec2 describe-instances --filters "Name=tag:Project,Values=minecraft" --query 'Reservations[*].Instances[0].PublicIpAddress' --output text | tr -d '\n'
deploy-world:
	rsync -avz -e "ssh -i ./data/aws-server-key.pem -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" --progress ./world.zip ec2-user@$(shell make get_ip):/home/ec2-user/mc-server/data/world.zip
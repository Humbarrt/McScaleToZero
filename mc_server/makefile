connect:
	ssh -i "./data/mc-server-key.pem" -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no ec2-user@$(shell make get_ip)
get_ip:
	@aws ec2 describe-instances --filters "Name=tag:Project,Values=minecraft" --query 'Reservations[*].Instances[0].PublicIpAddress' --output text | tr -d '\n'

# Assuming you're using a default VPC (no need to create one if not customized)
resource "aws_vpc" "mc_vpc" {
  cidr_block = "10.0.0.0/16"  # Only if creating your own VPC
  tags = {
    Name    = "mc-vpc"
    Project = "minecraft"
  }
}

# Tagging the default subnet (if using default VPC)
resource "aws_subnet" "mc_subnet" {
  vpc_id     = aws_vpc.mc_vpc.id 
  cidr_block = "10.0.1.0/24"  # Adjust to your network range
  tags = {
    Name    = "mc-subnet"
    Project = "minecraft"
  }
}

# Create a security group for the Minecraft server
resource "aws_security_group" "mc_sg" {
  name        = "mc-security-group"
  description = "Allow SSH and Minecraft traffic"
  
  tags = {
    Project = "minecraft"
  }

  # Allow SSH access
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"] # Change to restrict SSH access
  }

  # Allow Minecraft traffic on port 25565
  ingress {
    from_port   = 25565
    to_port     = 25565
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  # Allow all outbound traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}


# Create an EC2 instance for the Minecraft server with CloudWatch logging
resource "aws_instance" "mc_server" {
  ami           = "ami-0ecf75a98fe8519d7" # Amazon Linux 2023 (replace with the latest AMI, if you want)
  instance_type = "t3a.medium"
  key_name      = aws_key_pair.mc_ssh_key.key_name
  security_groups = [aws_security_group.mc_sg.name]
  iam_instance_profile = aws_iam_instance_profile.mc_instance_profile.name
  associate_public_ip_address = true # Auto-assign public IP

  # Define the root volume with 15GB of gp3 storage
  root_block_device {
    volume_size = 15
    volume_type = "gp3"
    delete_on_termination = true
    tags = {
      Project = "minecraft"
    }
  }

  tags = {
    Name = "mc-server"
    Project = "minecraft"
  }

  # Install CloudWatch agent and configure logging
  user_data = <<EOF
#!/bin/bash
# Install necessary packages
sudo yum update -y
sudo yum install -y docker

# Start Docker
sudo systemctl enable docker
sudo systemctl start docker

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

sudo mkdir -p /home/ec2-user/mc-server
# Write docker-compose.yml
sudo cat <<EOT > /home/ec2-user/mc-server/docker-compose.yml
  services:
    mc:
      container_name: mc-server
      image: itzg/minecraft-server:latest
      ports:
        - "25565:25565"
      environment:
        EULA: "TRUE"
        ENABLE_AUTOSTOP: "TRUE"
        AUTOSTOP_PKILL_USE_SUDO: "TRUE"
        AUTOSTOP_TIMEOUT_EST: "120"
        AUTOSTOP_TIMEOUT_INIT: "120"
        MEMORY: "3G"
        JVM_DD_OPTS: "java.net.preferIPv6Addresses=true"
      volumes:
        - /home/ec2-user/mc-server/data:/data
      restart: on-failure
      logging:
        driver: awslogs
        options:
          awslogs-group: "/mc-server/logs"
          awslogs-stream: "mc-server"
EOT

# Create mc-monitor script
sudo cat <<'EOT' > /home/ec2-user/mc-server/mc-monitor.sh
#!/bin/bash

cd /home/ec2-user/mc-server

/usr/local/bin/docker-compose up -d

while true; do
  sleep 5
  container_id=$(docker ps -aqf "name=mc-server")
  exit_code=$(docker inspect -f '{{.State.ExitCode}}' "$container_id")
  running=$(docker inspect -f '{{.State.Running}}' "$container_id")

  if [[ "$running" == "false" ]]; then
    echo "Container stopped. Waiting 10sec, before shutting down..."
    sleep 10
    running=$(docker inspect -f '{{.State.Running}}' "$container_id")
    if [[ "$running" == "true" ]]; then
      echo "Container is running again..."
      continue
    fi
    shutdown -h now
  fi
done
EOT

sudo chmod +x /home/ec2-user/mc-server/mc-monitor.sh

# Create systemd service
sudo cat <<EOT > /etc/systemd/system/mc-monitor.service
[Unit]
Description=Minecraft Container Monitor
After=network.target docker.service
Requires=docker.service

[Service]
Type=simple
ExecStart=/home/ec2-user/mc-server/mc-monitor.sh
Restart=always
User=root

[Install]
WantedBy=multi-user.target
EOT

# Enable and start service
sudo systemctl daemon-reexec
sudo systemctl daemon-reload
sudo systemctl enable mc-monitor
sudo systemctl start mc-monitor
EOF
}

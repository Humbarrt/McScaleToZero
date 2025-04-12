
# Generate an SSH key pair for the instance
resource "tls_private_key" "mc_ssh_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "mc_ssh_key" {
  key_name   = "mc-server-key"
  public_key = tls_private_key.mc_ssh_key.public_key_openssh
}

# Output the private key to a local file (ensure this is stored securely)
resource "local_file" "mc_private_key" {
  filename = "./data/mc-server-key.pem"
  content  = tls_private_key.mc_ssh_key.private_key_pem
}
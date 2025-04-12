
# Create a CloudWatch log group for storing logs
resource "aws_cloudwatch_log_group" "mc_log_group" {
  name = "/mc-server/logs"
  retention_in_days = 7 # Adjust as needed
  tags = {
    Project = "minecraft"
  }
}

# Create an IAM role to allow the EC2 instance to access CloudWatch
resource "aws_iam_role" "mc_server_role" {
  name = "mc-server-role"

  tags = {
    Project = "minecraft"
  }

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

# Define IAM policy to grant CloudWatch permissions to the instance
resource "aws_iam_role_policy" "mc_policy" {
  name = "mc-server-policy"
  role = aws_iam_role.mc_server_role.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": [
        "${aws_cloudwatch_log_group.mc_log_group.arn}",
        "${aws_cloudwatch_log_group.mc_log_group.arn}:*"
      ]
    }
  ]
}
EOF
}

# Attach the IAM role to an instance profile
resource "aws_iam_instance_profile" "mc_instance_profile" {
  name = "mc-instance-profile"
  role = aws_iam_role.mc_server_role.name
}
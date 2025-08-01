# Create an IAM user
resource "aws_iam_user" "mc_manager_user" {
  name = "mc-manager-user"
}

# Create an access key for the IAM user
resource "aws_iam_access_key" "mc_manager_key" {
  user = aws_iam_user.mc_manager_user.name
}

# Create an IAM policy allowing start/stop and cost viewing
data "aws_iam_policy_document" "mc_manager_policy_doc" {
  statement {
    sid    = "EC2DescribeInstances"
    effect = "Allow"
    actions = [
      "ec2:DescribeInstances"
    ]
    resources = ["*"]
  }
  statement {
    sid    = "EC2Control"
    effect = "Allow"
    actions = [
      "ec2:StartInstances",
      "ec2:StopInstances"
    ]
    resources = [
        aws_instance.mc_server.arn
    ]
  }

  statement {
    sid    = "ViewBilling"
    effect = "Allow"
    actions = [
      "ce:ListCostAllocationTags",
      "ce:ListCostAllocationTagBackfillHistory",
      "ce:DescribeCostCategoryDefinition",
      "ce:GetCostAndUsageWithResources",
      "ce:GetCostForecast",
      "ce:GetCostAndUsage",
      "ce:DescribeReport",
      "ce:GetTags",
      "ce:StartCostAllocationTagBackfill",
      "ce:GetCostCategories",
      "ce:GetDimensionValues"
    ]
    resources = ["*"]
  }

  statement {
    sid    = "CloudWatchLogsAccess"
    effect = "Allow"
    actions = [
      "logs:GetLogEvents",
      "logs:DescribeLogStreams",
      "logs:DescribeLogGroups",
      "logs:StartQuery",
      "logs:GetQueryResults",
      "logs:FilterLogEvents"
    ]
    resources = [
      "${aws_cloudwatch_log_group.mc_log_group.arn}",
      "${aws_cloudwatch_log_group.mc_log_group.arn}:*"
    ]
  }
}

# Attach the policy to the user
resource "aws_iam_user_policy" "mc_manager_user_policy" {
  name   = "mc-access-policy"
  user   = aws_iam_user.mc_manager_user.name
  policy = data.aws_iam_policy_document.mc_manager_policy_doc.json
}

# Output the access key and secret
resource "local_file" "mc_manager_user_access_key_id" {
  filename = "./data/aws-access-key-id.txt"
  content  = aws_iam_access_key.mc_manager_key.id
}

resource "local_file" "mc_manager_user_secret_access_key" {
  filename = "./data/aws-secret-access-key.txt"
  content  = aws_iam_access_key.mc_manager_key.secret
}

data "aws_region" "current" {}

# Output the AWS Region as local file
resource "local_file" "aws_region" {
  filename = "./data/aws-region.txt"
  content  = data.aws_region.current.name
}
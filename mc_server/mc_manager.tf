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
    sid    = "EC2Control"
    effect = "Allow"
    actions = [
      "ec2:StartInstances",
      "ec2:StopInstances",
      "ec2:DescribeInstances"
    ]
    resources = ["*"]
    condition {
      test     = "StringEquals"
      variable = "aws:RequestTag/Project"
      values   = ["minecraft"]
    }
  }

  statement {
    sid    = "ViewBilling"
    effect = "Allow"
    actions = [
      "ce:GetCostAndUsage",
      "ce:GetCostForecast",
      "ce:GetDimensionValues",
      "ce:GetReservationUtilization",
      "ce:GetRightsizingRecommendation"
    ]
    resources = ["*"]
    condition {
      test     = "StringEquals"
      variable = "aws:RequestTag/Project"
      values   = ["minecraft"]
    }
  }

  statement {
    sid    = "CloudWatchLogsAccess"
    effect = "Allow"
    actions = [
      "logs:DescribeLogGroups",
      "logs:DescribeLogStreams",
      "logs:GetLogEvents"
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
  filename = "./data/mc-access-key-id.txt"
  content  = aws_iam_access_key.mc_manager_key.id
}

resource "local_file" "mc_manager_user_secret_access_key" {
  filename = "./data/mc-secret-access-key.txt"
  content  = aws_iam_access_key.mc_manager_key.secret
}
# Define the AWS provider and use the default region from the user's AWS profile
provider "aws" {
  # No region specified, uses default from AWS CLI configuration (~/.aws/config)
}

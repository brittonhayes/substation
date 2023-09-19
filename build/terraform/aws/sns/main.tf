resource "aws_sns_topic" "topic" {
  name                        = var.config.name
  kms_master_key_id           = var.kms.id
  fifo_topic                  = endswith(var.config.name, ".fifo") ? true : false
  content_based_deduplication = endswith(var.config.name, ".fifo") ? true : false

  tags = var.tags
}

# Applies the policy to each role in the access list.
resource "aws_iam_role_policy_attachment" "access" {
  count      = length(var.access)
  role       = var.access[count.index]
  policy_arn = aws_iam_policy.access.arn
}

resource "aws_iam_policy" "access" {
  name        = "sub-sns-${var.config.name}"
  description = "Policy for the ${var.config.name} SNS topic."
  policy      = data.aws_iam_policy_document.access.json
}

data "aws_iam_policy_document" "access" {
  statement {
    effect = "Allow"
    actions = [
      "kms:Decrypt",
      "kms:GenerateDataKey"
    ]

    resources = [
      var.kms.arn,
    ]
  }

  statement {
    effect = "Allow"
    actions = [
      "sns:Publish",
    ]

    resources = [
      aws_sns_topic.topic.arn,
    ]
  }
}

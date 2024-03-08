data "aws_iam_policy_document" "spotify_bucket_iam_policy" {
  statement {
    sid = "PublicReadGetObject"
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
    effect    = "Allow"
    actions   = ["s3:GetObject"]
    resources = ["${aws_s3_bucket.spotify_bucket.arn}/*"]
  }
}

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

## S3 Configuration ##

resource "aws_s3_bucket" "spotify_bucket" {
  bucket        = "spotify-clone-bucket"
  force_destroy = true
}

resource "aws_s3_bucket_ownership_controls" "spotify_bucket_ownership" {
  bucket = aws_s3_bucket.spotify_bucket.id
  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

resource "aws_s3_bucket_public_access_block" "spotify_bucket_public_access" {
  bucket = aws_s3_bucket.spotify_bucket.id

  block_public_acls       = false
  block_public_policy     = false
  ignore_public_acls      = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_acl" "spotify_bucket_acl" {
  bucket = aws_s3_bucket.spotify_bucket.id
  acl    = "private"
  depends_on = [
    aws_s3_bucket_ownership_controls.spotify_bucket_ownership,
    aws_s3_bucket_public_access_block.spotify_bucket_public_access
  ]
}

resource "aws_s3_bucket_cors_configuration" "spotify_bucket_cors" {
  bucket = aws_s3_bucket.spotify_bucket.id
  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET"]
    allowed_origins = ["*"]
    expose_headers = [
      "x-amz-server-side-encryption",
      "x-amz-request-id",
      "x-amz-id-2"
    ]
    max_age_seconds = 3600
  }
}

## CloudFront Configuration ##

resource "aws_cloudfront_origin_access_identity" "spotify_oai" {
  comment = "Spotify CDN Distribution OAI"
}

resource "aws_cloudfront_distribution" "spotify_s3_distribution" {
  origin {
    domain_name = aws_s3_bucket.spotify_bucket.bucket_regional_domain_name
    origin_id   = "S3-${aws_s3_bucket.spotify_bucket.id}"

    s3_origin_config {
      origin_access_identity = aws_cloudfront_origin_access_identity.spotify_oai.cloudfront_access_identity_path
    }
  }

  enabled         = true
  is_ipv6_enabled = true

  default_cache_behavior {
    allowed_methods  = ["GET", "HEAD", "OPTIONS"]
    cached_methods   = ["HEAD", "GET"]
    target_origin_id = "S3-${aws_s3_bucket.spotify_bucket.id}"

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
    viewer_protocol_policy = "redirect-to-https"
  }

  restrictions {
    geo_restriction {
      restriction_type = "whitelist"
      locations        = ["IN"] # Restrict CDN distribution to india only
    }
  }

  viewer_certificate {
    cloudfront_default_certificate = true
  }
}

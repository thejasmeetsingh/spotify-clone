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
  ignore_public_acls      = false
  block_public_policy     = false
  restrict_public_buckets = false
}

resource "aws_s3_bucket_policy" "spotify_bucket_policy" {
  bucket = aws_s3_bucket.spotify_bucket.id
  policy = data.aws_iam_policy_document.spotify_bucket_iam_policy.json
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
    allowed_methods  = ["HEAD", "GET"]
    cached_methods   = ["HEAD", "GET"]
    target_origin_id = "S3-${aws_s3_bucket.spotify_bucket.id}"

    forwarded_values {
      query_string = false

      cookies {
        forward = "none"
      }
    }

    viewer_protocol_policy = "allow-all"
    min_ttl                = 0
    default_ttl            = 3600
    max_ttl                = 86400
  }

  ordered_cache_behavior {
    path_pattern     = "/cdn/*"
    allowed_methods  = ["GET", "HEAD", "OPTIONS"]
    cached_methods   = ["GET", "HEAD"]
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
    compress               = true
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

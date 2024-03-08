terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

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

resource "alicloud_oss_bucket" "bucket-acl" {
  bucket = "bucket-170309-lifecycle"
  acl    = "private"
}


resource "alicloud_oss_bucket_lifecycle" "bucket-lifecycle" {
  bucket = alicloud_oss_bucket.bucket-acl.bucket

  life_cycle_id = "rule-tags"
  prefix = "path3/"
  enabled = true

  tags = {
    key = "value1"
    value = "value2"
  }
  abort_multipart_upload {
    days = 100
  }

}
resource "alicloud_oss_bucket_lifecycle" "bucket-lifecycle2" {
  bucket = alicloud_oss_bucket.bucket-acl.bucket

  life_cycle_id = "rule-tags"
  prefix = "path3/"
  enabled = true

  tags = {
    key = "value1"
    value = "value2"
  }
  abort_multipart_upload {
    days = 100
  }

}

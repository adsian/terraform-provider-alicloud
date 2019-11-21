---
layout: "alicloud"
page_title: "Alicloud: alicloud_oss_bucket"
sidebar_current: "docs-alicloud-resource-oss-bucket"
description: |-
  Provides a resource to create a oss bucket.
---

# alicloud\_oss\_bucket\_lifecycle

Provides a resource to create lifecycle rules for a bucket. After lifecycle rules are configured for a bucket, OSS automatically deletes the objects that conform to the lifecycle rules on a regular basis.

-> **NOTE:** If a lifecycle rule is configured for the bucket, this operation overwrites the previous lifecycle rule.


## Example Usage

Set lifecycle expiration with tags

```
resource "alicloud_oss_bucket_lifecycle" "bucket-lifecycle" {
  bucket = "bucket-170309-lifecycle"
  acl    = "public-read"

  lifecycle_rule {
    id      = "rule-days"
    prefix  = "path1/"
    enabled = true

    tags {
      key = "value1"
      value = "value2"
    }

    tags {
       key = "value1"
       value = "value2"
    }

    expiration {
      days = 365
    }
  }

  lifecycle_rule {
    id      = "rule-date"
    prefix  = "path2/"
    enabled = true

    tags {
      key = "value1"
      value = "value2"
    }

    tags {
       key = "value1"
       value = "value2"
    }

    expiration {
      date = "2018-01-12"
    }
  }
}
```

Set lifecycle rule abort_multipart_upload

```
resource "alicloud_oss_bucket_lifecycle" "bucket-lifecycle" {
  bucket = "bucket-170309-lifecycle"
  acl    = "public-read"

  lifecycle_rule {
    id      = "rule-days"
    prefix  = "path1/"
    enabled = true

    abort_multipart_upload {
      days = 3
    }
  }

  lifecycle_rule {
    id      = "rule-date"
    prefix  = "path2/"
    enabled = true

    abort_multipart_upload {
      created_before_date = "2015-11-11"
    }
  }
}
```

Set lifecycle transitions

```
resource "alicloud_oss_bucket_lifecycle" "bucket-lifecycle" {
  bucket = "bucket-170309-lifecycle"
  acl    = "public-read"

  lifecycle_rule {
    id      = "rule-days-transition"
    prefix  = "path3/"
    enabled = true

    transitions {
        days =         "3"
        storage_class= "IA"
    }

    transitions {
        days=         "30"
        storage_class= "Archive"
    }
  }
}

resource "alicloud_oss_bucket_lifecycle" "bucket-lifecycle" {
  bucket = "bucket-170309-lifecycle"
  acl    = "public-read"

  lifecycle_rule {
    id      = "rule-days-transition"
    prefix  = "path3/"
    enabled = true

    transitions {
      created_before_date = "2020-11-11"
      storage_class = "IA"
    }
    transitions {
      created_before_date = "2021-11-11"
      storage_class = "Archive"
    }
  }
}
```


## Argument Reference

The following arguments are supported:

* `bucket` - (Optional, ForceNew) The name of the bucket.
* `acl` - (Optional) The [canned ACL](https://www.alibabacloud.com/help/doc-detail/31898.htm) to apply. Defaults to "private".
* `lifecycle_rule` - (Optional) A configuration of [object lifecycle management](https://www.alibabacloud.com/help/doc-detail/31904.htm) (documented below).

#### Block lifecycle_rule

The lifecycle_rule object supports the following:

* `id` - (Optional) Unique identifier for the rule. If omitted, OSS bucket will assign a unique name.
* `prefix` - (Required) Object key prefix identifying one or more objects to which the rule applies.
* `enabled` - (Required, Type: bool) Specifies lifecycle rule status.
* `tags` - (Required, Type: set, Available in 1.63.0+) Specifies the object tag applicable to a rule. Multiple tags are supported (documented below).
* `expiration` - (Optional, Type: set, Available in 1.63.0+) Specifies the expiration attributes of the lifecycle rules for the object (documented below).
* `abort_multipart_upload` - (Optional, Type: set, Available in 1.63.0+) Specifies the expiration attributes of the multipart upload tasks that are not complete (documented below).
* `transitions` - (Optional, Type: set, Available in 1.63.0+) Specifies the time when an object is converted to the IA or Archive storage class during a valid life cycle (documented below).

#### Block tags

The lifecycle_rule tags object supports the following:

* `key` - (Type: string) Indicates the tag key.
* `value` - (Type: string) Indicates the tag value.

`NOTE`: The "key" and "value" were required if their father node tags already set.

#### Block expiration

The lifecycle_rule expiration object supports the following:

* `date` - (Optional) Specifies the date after which you want the corresponding action to take effect. The value obeys ISO8601 format like `2017-03-09`. "date" is an absolute expiration time: The expiration time in date is not recommended.
* `created_before_date` - (Optional) Specifies the time before which the rules take effect. The date must conform to the ISO8601 format and always be UTC 00:00. For example: 2002-10-11T00:00:00.000Z indicates that objects updated before 2002-10-11T00:00:00.000Z are deleted or converted to another storage class, and objects updated after this time (including this time) are not deleted or converted.
* `days` - (Optional, Type: int) Specifies the number of days after object creation when the specific rule action takes effect.

`NOTE`: One and only one of "date", "created_before_date" and "days" can be specified in one expiration configuration.

#### Block abort_multipart_upload

The lifecycle_rule abort_multipart_upload object supports the following:

* `created_before_date` - (Optional) Specifies the time before which the rules take effect. The date must conform to the ISO8601 format and always be UTC 00:00. For example: 2002-10-11T00:00:00.000Z indicates that objects updated before 2002-10-11T00:00:00.000Z are deleted or converted to another storage class, and objects updated after this time (including this time) are not deleted or converted.
* `days` - (Optional, Type: int) Specifies how many days after the object is updated for the last time until the rules take effect.

`NOTE`: One and only one of "created_before_date" and "days" can be specified in one abort_multipart_upload configuration.

#### Block transitions

The lifecycle_rule transitions object supports the following:

* `created_before_date` - (Optional) Specifies the time before which the rules take effect. The date must conform to the ISO8601 format and always be UTC 00:00. For example: 2002-10-11T00:00:00.000Z indicates that objects updated before 2002-10-11T00:00:00.000Z are deleted or converted to another storage class, and objects updated after this time (including this time) are not deleted or converted.
* `days` - (Optional, Type: int) Specifies the number of days after object creation when the specific rule action takes effect.
* `storage_class` - (Required) Specifies the storage class that objects that conform to the rule are converted into. The storage class of the objects in a bucket of the IA storage class can be converted into Archive but cannot be converted into Standard. Values: `IA`, `Archive`, `Standard`. 

`NOTE`: One and only one of "created_before_date" and "days" can be specified in one transition configuration.

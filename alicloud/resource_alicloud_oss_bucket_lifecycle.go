package alicloud

import (
	"bytes"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/terraform-providers/terraform-provider-alicloud/alicloud/connectivity"
	"time"
)

// bucket lifecycle example
// https://github.com/aliyun/aliyun-oss-go-sdk/blob/master/sample/bucket_lifecycle.go

func resourceAlicloudOssBucketLifeCycle() *schema.Resource {
	return &schema.Resource{
		Create: resourceAlicloudOssBucketLifeCycleCreate,
		Read:   resourceAlicloudOssBucketLifeCycleRead,
		Update: resourceAlicloudOssBucketLifeCycleUpdate,
		Delete: resourceAlicloudOssBucketLifeCycleDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateOssBucketName,
				Default:      resource.PrefixedUniqueId("tf-oss-bucket-"),
			},
			"lifecycle_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validateOssBucketLifecycleRuleId,
			},
			"prefix": {
				Type:     schema.TypeString,
				Required: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			// Lifecycle rules can be configured with specific tags
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			// the rule's expiration property
			"expiration": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      lexpirationHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Absolute expiration time: The expiration time in date, not recommended
						"date": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateOssBucketDateTimestamp,
							// ConflictsWith: []string{"created_before_date", "days"},
						},
						// objects created before the date will be expired
						"created_before_date": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateOssBucketDateTimestamp,
							// ConflictsWith: []string{"days", "date"},
						},
						// Relative expiration time: The expiration time in days after the last modified time
						"days": {
							Type:     schema.TypeInt,
							Optional: true,
							// ConflictsWith: []string{"date", "created_before_date"},
						},
					},
				},
			},

			// the rule's abort multipart upload property
			"abort_multipart_upload": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      abortMultipartUploadHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// Relative expiration time: The expiration time in days after the last modified time
						"created_before_date": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateOssBucketDateTimestamp,
							// ConflictsWith: []string{"days"},
						},
						// objects created before the date will be expired
						"days": {
							Type:     schema.TypeInt,
							Optional: true,
							// ConflictsWith: []string{"created_before_date"},
						},
					},
				},
			},

			// the rule's transition property
			"transitions": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      ltransitionsHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// objects created before the date will be expired
						"created_before_date": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validateOssBucketDateTimestamp,
							// ConflictsWith: []string{"days"},
						},
						// Relative transition time: The transition time in days after the last modified time
						"days": {
							Type:     schema.TypeInt,
							Optional: true,
							// ConflictsWith: []string{"created_before_date"},
						},
						// objects created before the date will be expired
						"storage_class": {
							Type:     schema.TypeString,
							Default:  oss.StorageStandard,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validateAllowedStringValue([]string{
								string(oss.StorageStandard),
								string(oss.StorageIA),
								string(oss.StorageArchive),
							}),
						},
					},
				},
			},
		},
	}
}

func resourceAlicloudOssBucketLifeCycleCreate(d *schema.ResourceData, meta interface{}) error {
	// todo

	client := meta.(*connectivity.AliyunClient)
	request := map[string]string{"bucketName": d.Get("bucket").(string)}
	var requestInfo *oss.Client
	raw, err := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
		requestInfo = ossClient
		return ossClient.IsBucketExist(request["bucketName"])
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, "alicloud_oss_bucket", "IsBucketExist", AliyunOssGoSdk)
	}
	addDebug("IsBucketExist", raw, requestInfo, request)
	isExist, _ := raw.(bool)
	if isExist {
		return WrapError(Error("[ERROR] The specified bucket name: %#v is not available. The bucket namespace is shared by all users of the OSS system. Please select a different name and try again.", request["bucketName"]))
	}
	type Request struct {
		BucketName string
		Option     oss.Option
	}

	return resourceAlicloudOssBucketLifeCycleUpdate(d, meta)
}

func resourceAlicloudOssBucketLifeCycleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ossService := OssService{client}
	object, err := ossService.DescribeOssBucket(d.Id())
	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return WrapError(err)
	}

	d.Set("bucket", d.Id())
	d.Set("acl", object.BucketInfo.ACL)

	request := map[string]string{"bucketName": d.Id()}
	var requestInfo *oss.Client

	// d.SetId(request.DiskId + ":" + request.InstanceId) ???

	raw, err := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
		return ossClient.GetBucketLifecycle(d.Id())
	})
	if err != nil && !ossNotFoundError(err) {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "GetBucketLifecycle", AliyunOssGoSdk)
	}
	addDebug("GetBucketLifecycle", raw, requestInfo, request)

	// Read bucket tags
	raw, err = client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
		return ossClient.GetBucketTagging(d.Id())
	})

	// Read the lifecycle rule configuration
	lrules := make([]map[string]interface{}, 0)
	lifecycle, _ := raw.(oss.GetBucketLifecycleResult)
	for _, lifecycleRule := range lifecycle.Rules {
		rule := make(map[string]interface{})
		rule["id"] = lifecycleRule.ID
		rule["prefix"] = lifecycleRule.Prefix
		if LifecycleRuleStatus(lifecycleRule.Status) == ExpirationStatusEnabled {
			rule["enabled"] = true
		} else {
			rule["enabled"] = false
		}

		// tags
		if lifecycleRule.Tags != nil {
			eTotal := make([]map[string]interface{}, 0)
			for _, tag := range lifecycleRule.Tags {
				e := make(map[string]interface{})
				e["key"] = tag.Key
				e["value"] = tag.Value

				eTotal = append(eTotal, e)
			}
			rule["tags"] = schema.NewSet(tagsHash, []interface{}{eTotal})
		}

		// expiration
		if lifecycleRule.Expiration != nil {
			e := make(map[string]interface{})

			if lifecycleRule.Expiration.Date != "" {
				t, err := time.Parse("2006-01-02T15:04:05.000Z", lifecycleRule.Expiration.Date)
				if err != nil {
					return WrapError(err)
				}
				e["date"] = t.Format("2006-01-02")
			}

			if lifecycleRule.Expiration.CreatedBeforeDate != "" {
				t, err := time.Parse("2006-01-02T15:04:05.000Z", lifecycleRule.Expiration.CreatedBeforeDate)
				if err != nil {
					return WrapError(err)
				}
				e["create_before_date"] = t.Format("2006-01-02")
			}
			e["days"] = int(lifecycleRule.Expiration.Days)
			rule["expiration"] = schema.NewSet(lexpirationHash, []interface{}{e})
		}

		// AbortMultiPartUpload
		if lifecycleRule.AbortMultipartUpload != nil {
			e := make(map[string]interface{})

			if lifecycleRule.AbortMultipartUpload.Days != 0 {
				e["days"] = lifecycleRule.AbortMultipartUpload.Days
			}

			if lifecycleRule.AbortMultipartUpload.CreatedBeforeDate != "" {
				t, err := time.Parse("2006-01-02T15:04:05.000Z", lifecycleRule.AbortMultipartUpload.CreatedBeforeDate)
				if err != nil {
					return WrapError(err)
				}
				e["create_before_date"] = t.Format("2006-01-02")
			}
			rule["abort_multipart_upload"] = schema.NewSet(abortMultipartUploadHash, []interface{}{e})
		}

		// Transitions
		if len(lifecycleRule.Transitions) != 0 {
			var eSli []interface{}
			for _, transition := range lifecycleRule.Transitions {
				e := make(map[string]interface{})
				if transition.CreatedBeforeDate != "" {
					t, err := time.Parse("2006-01-02T15:04:05.000Z", transition.CreatedBeforeDate)
					if err != nil {
						return WrapError(err)
					}
					e["created_before_date"] = t.Format("2006-01-02")
				}
				e["days"] = transition.Days
				e["storage_class"] = string(transition.StorageClass)
				eSli = append(eSli, e)
			}
			rule["transitions"] = schema.NewSet(ltransitionsHash, eSli)
		}

		lrules = append(lrules, rule)
	}

	if err := d.Set("lifecycle_rule", lrules); err != nil {
		return WrapError(err)
	}

	return nil
}

func resourceAlicloudOssBucketLifeCycleUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	d.Partial(true)

	if d.HasChange("lifecycle_id") {
		if err := resourceAlicloudOssBucketLifeCycleUpdatePartial(client, d); err != nil {
			return WrapError(err)
		}
		d.SetPartial("lifecycle_id")
	}

	return resourceAlicloudOssBucketLifeCycleRead(d, meta)
}

func resourceAlicloudOssBucketLifeCycleUpdatePartial(client *connectivity.AliyunClient, d *schema.ResourceData) error {
	bucket := d.Id()
	lifecycleRule := d.Get("lifecycle_id").(interface{})
	var requestInfo *oss.Client

	if lifecycleRule == nil {
		raw, err := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
			requestInfo = ossClient
			return nil, ossClient.DeleteBucketLifecycle(bucket)
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteBucketLifecycle", AliyunOssGoSdk)

		}
		addDebug("DeleteBucketLifecycle", raw, requestInfo, map[string]interface{}{
			"bucketName": bucket,
		})
		return nil
	}

	rules := make([]oss.LifecycleRule, 0, len(lifecycleRules))

	for i, lifecycleRule := range lifecycleRules {
		r := lifecycleRule.(map[string]interface{})

		rule := oss.LifecycleRule{
			Prefix: r["prefix"].(string),
		}

		// id
		if val, ok := r["id"].(string); ok && val != "" {
			rule.ID = val
		}

		// enabled
		if val, ok := r["enabled"].(bool); ok && val {
			rule.Status = string(ExpirationStatusEnabled)
		} else {
			rule.Status = string(ExpirationStatusDisabled)
		}

		// tags
		tags := d.Get(fmt.Sprintf("lifecycle_rule.%d.tags", i)).(*schema.Set).List()
		if len(tags) > 0 {
			for _, tag := range tags {
				//rule.Tags = append(rule.Tags, tag.(oss.Tag))
				valTag := oss.Tag{}
				valKey := tag.(map[string]interface{})["key"].(string)
				valValue := tag.(map[string]interface{})["value"].(string)

				if valKey != "" {
					valTag.Key = valKey
				}

				if valValue != "" {
					valTag.Value = valValue
				}

				rule.Tags = append(rule.Tags, valTag)
			}
		}

		// expiration
		expiration := d.Get(fmt.Sprintf("lifecycle_rule.%d.expiration", i)).(*schema.Set).List()
		if len(expiration) > 0 {
			e := expiration[0].(map[string]interface{})
			i := oss.LifecycleExpiration{}
			valCreatedBeforeDate, _ := e["create_before_date"].(string)
			valDate, _ := e["date"].(string)
			valDays, _ := e["days"].(int)

			//if (valCreatedBeforeDate != "" && valDays > 0) || (valCreatedBeforeDate == "" && valDays <= 0) {
			//	return WrapError(Error("'date' conflicts with 'days'. One and only one of them can be specified in one expiration configuration."))
			//}

			if valCreatedBeforeDate != "" {
				i.CreatedBeforeDate = fmt.Sprintf("%sT00:00:00.000Z", valCreatedBeforeDate)
			}

			if valDate != "" {
				i.Date = fmt.Sprintf("%sT00:00:00.000Z", valDate)
			}

			if valDays > 0 {
				i.Days = valDays
			}
			rule.Expiration = &i
		}

		// abortMultipartUpload
		abortMultipartUpload := d.Get(fmt.Sprintf("lifecycle_rule.%d.abort_multipart_upload", i)).(*schema.Set).List()
		if len(abortMultipartUpload) > 0 {
			e := abortMultipartUpload[0].(map[string]interface{})
			i := oss.LifecycleAbortMultipartUpload{}
			valCreatedBeforeDate, _ := e["create_before_date"].(string)
			valDays, _ := e["days"].(int)

			//if (valCreatedBeforeDate != "" && valDays > 0) || (valCreatedBeforeDate == "" && valDays <= 0) {
			//	return WrapError(Error("'date' conflicts with 'days'. One and only one of them can be specified in one expiration configuration."))
			//}

			if valCreatedBeforeDate != "" {
				i.CreatedBeforeDate = fmt.Sprintf("%sT00:00:00.000Z", valCreatedBeforeDate)
			}

			if valDays > 0 {
				i.Days = valDays
			}

			rule.AbortMultipartUpload = &i
		}

		// transitions
		transitions := d.Get(fmt.Sprintf("lifecycle_rule.%d.transitions", i)).(*schema.Set).List()
		if len(transitions) > 0 {
			for _, transition := range transitions {
				i := oss.LifecycleTransition{}

				valCreatedBeforeDate := transition.(map[string]interface{})["created_before_date"].(string)
				valDays := transition.(map[string]interface{})["days"].(int)
				valStorageClass := transition.(map[string]interface{})["storage_class"].(string)

				//if (valCreatedBeforeDate != "" && valDays > 0) || (valCreatedBeforeDate == "" && valDays <= 0) || (valStorageClass == "") {
				//	return WrapError(Error("'CreatedBeforeDate' conflicts with 'Days'. One and only one of them can be specified in one transition configuration. 'storage_class' must be set."))
				//}

				if valCreatedBeforeDate != "" {
					i.CreatedBeforeDate = fmt.Sprintf("%sT00:00:00.000Z", valCreatedBeforeDate)
				}
				if valDays > 0 {
					i.Days = valDays
				}

				if valStorageClass != "" {
					i.StorageClass = oss.StorageClassType(valStorageClass)
				}
				rule.Transitions = append(rule.Transitions, i)
			}
		}

		rules = append(rules, rule)
	}

	raw, err := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
		requestInfo = ossClient
		return nil, ossClient.SetBucketLifecycle(bucket, rules)
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "SetBucketLifecycle", AliyunOssGoSdk)
	}
	addDebug("SetBucketLifecycle", raw, requestInfo, map[string]interface{}{
		"bucketName": bucket,
		"rules":      rules,
	})

	return nil
}

func resourceAlicloudOssBucketLifeCycleDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)
	ossService := OssService{client}
	var requestInfo *oss.Client
	raw, err := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
		requestInfo = ossClient
		return ossClient.IsBucketExist(d.Id())
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "IsBucketExist", AliyunOssGoSdk)
	}
	addDebug("IsBucketExist", raw, requestInfo, map[string]string{"bucketName": d.Id()})

	exist, _ := raw.(bool)
	if !exist {
		return nil
	}

	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		raw, er := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
			return nil, ossClient.DeleteBucketLifecycle(d.Id())
		})

		if er != nil {
			if IsExceptedError(er, "NoSuchBucket") {
				return resource.NonRetryableError(er)
			}

			if IsExceptedError(er, "AccessDenied") {
				return resource.NonRetryableError(er)
			}
		}

		addDebug("DeleteBucketLifeCycle", raw, requestInfo, map[string]string{"bucketName": d.Id()})
		return nil
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "DeleteBucketLifeCycle", AliyunOssGoSdk)
	}
	return WrapError(ossService.WaitForOssBucket(d.Id(), Changing, DefaultTimeoutMedium))

	return nil
}

func tagsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["key"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["value"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	return hashcode.String(buf.String())
}

func ltransitionsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["created_before_date"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["storage_class"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["days"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	return hashcode.String(buf.String())
}

func abortMultipartUploadHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["created_before_date"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["days"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	return hashcode.String(buf.String())
}

func lexpirationHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["date"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["days"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	return hashcode.String(buf.String())
}

package alicloud

import (
	"bytes"
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/terraform-providers/terraform-provider-alicloud/alicloud/connectivity"
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
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(3, 63),
			},
			"life_cycle_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
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
			"tags": tagsSchema(),
			// the rule's expiration property
			"expiration": {
				Type:     schema.TypeSet,
				Optional: true,
				//Set:      lexpirationHash,
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
				//Set:      abortMultipartUploadHash,
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
				//Set:      ltransitionsHash,
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
							ValidateFunc: validation.StringInSlice([]string{
								string(oss.StorageStandard),
								string(oss.StorageIA),
								string(oss.StorageArchive),
							}, false),
						},
					},
				},
			},
		},
	}
}

func resourceAlicloudOssBucketLifeCycleCreate(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*connectivity.AliyunClient)
	// var ossClient *oss.Client
	bucket := d.Id()
	var requestInfo *oss.Client
	//request := map[string]string{"bucketName": d.Id()}

	// Read the lifecycle rule configuration
	// lifecycle rule is a overriding attribute
	// it should get the rules remotely firstly, after committed comparing and merging
	// then push the new life cycle rules back

	newLifeCycleRule := oss.LifecycleRule{
		ID:                   d.Get("life_cycle_id").(string),
		Prefix:               d.Get("prefix").(string),
		Status:               d.Get("enabled").(string),
		Tags:                 d.Get("tags").([]oss.Tag),
		Expiration:           d.Get("expiration").(*oss.LifecycleExpiration),
		AbortMultipartUpload: d.Get("abort_multipart_upload").(*oss.LifecycleAbortMultipartUpload),
		Transitions:          d.Get("transitions").([]oss.LifecycleTransition),
	}

	existedLifeCycle, err := readTheLifeCycleRuleConfiguration(d, meta)
	totalLifeCycle := make([]oss.LifecycleRule, 0)
	totalLifeCycle = append(existedLifeCycle, newLifeCycleRule)

	if err != nil {
		return err
	}

	lifeCycleId := d.Get("life_cycle_id")

	if lifeCycleId == "" {
		// push firstly, then set take back the life_cycle_id and set locally

		raw, err := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
			return nil, ossClient.SetBucketLifecycle(bucket, totalLifeCycle)
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "SetBucketLifecycle", AliyunOssGoSdk)
		}
		addDebug("SetBucketLifecycle", raw, requestInfo, map[string]interface{}{
			"bucketName": bucket,
			"rules":      totalLifeCycle,
		})

		pullBackLifeCycles, err := readTheLifeCycleRuleConfiguration(d, meta)
		for _, ruleE := range existedLifeCycle {
			for _, ruleP := range pullBackLifeCycles {
				// find the return back id
				if ruleE.ID != ruleP.ID {
					// newLifeCycleRule.ID = ruleB.ID
					_ = d.Set("life_cycle_id", ruleP.ID)
				}
			}
		}
	} else {
		_ = d.Set("life_cycle_id", lifeCycleId)

		raw, err := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
			return nil, ossClient.SetBucketLifecycle(bucket, totalLifeCycle)
		})
		if err != nil {
			return WrapErrorf(err, DefaultErrorMsg, d.Id(), "SetBucketLifecycle", AliyunOssGoSdk)
		}
		addDebug("SetBucketLifecycle", raw, requestInfo, map[string]interface{}{
			"bucketName": bucket,
			"rules":      totalLifeCycle,
		})
	}

	return resourceAlicloudOssBucketLifeCycleRead(d, meta)
}

func resourceAlicloudOssBucketLifeCycleRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*connectivity.AliyunClient)

	_ = d.Set("bucket", d.Id())

	request := map[string]string{"bucket": d.Id()}
	var requestInfo *oss.Client

	// using api GetBucketLifecycle to get lifeCycles

	raw, err := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
		return ossClient.GetBucketLifecycle(d.Id())
	})
	if err != nil && !ossNotFoundError(err) {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "GetBucketLifecycle", AliyunOssGoSdk)
	}
	addDebug("GetBucketLifecycle", raw, requestInfo, request)

	lifeCycleRuleId := d.Get("life_cycle_id")
	lifecycle, _ := raw.(oss.GetBucketLifecycleResult)
	for _, lifeCycleRule := range lifecycle.Rules {
		if lifeCycleRule.ID == lifeCycleRuleId {
			_ = d.Set("life_cycle_id", lifeCycleRule.ID)
			_ = d.Set("prefix", lifeCycleRule.Prefix)

			if LifecycleRuleStatus(lifeCycleRule.Status) == ExpirationStatusEnabled {
				_ = d.Set("enabled", true)
			} else {
				_ = d.Set("enabled", false)
			}

			_ = d.Set("tags", lifeCycleRule.Tags)
			_ = d.Set("expiration", lifeCycleRule.Expiration)
			_ = d.Set("abort_multipart_upload", lifeCycleRule.AbortMultipartUpload)
			_ = d.Set("transitions", lifeCycleRule.Transitions)
		}
	}

	return nil
}

func resourceAlicloudOssBucketLifeCycleUpdate(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*connectivity.AliyunClient)
	// var ossClient *oss.Client
	bucket := d.Id()
	var requestInfo *oss.Client
	// Read the lifecycle rule configuration
	// lifecycle rule is a overriding attribute
	// it should get the rules remotely firstly, after committed comparing and merging
	// then push the new life cycle rules back

	lifeCycleRuleId := d.Get("life_cycle_id")
	existedLifeCycle, err := readTheLifeCycleRuleConfiguration(d, meta)

	if err != nil {
		return err
	}

	for _, rule := range existedLifeCycle {
		if lifeCycleRuleId == rule.ID {
			newLifeCycleRule := oss.LifecycleRule{
				ID:                   d.Get("life_cycle_id").(string),
				Prefix:               d.Get("prefix").(string),
				Status:               d.Get("enabled").(string),
				Tags:                 d.Get("tags").([]oss.Tag),
				Expiration:           d.Get("expiration").(*oss.LifecycleExpiration),
				AbortMultipartUpload: d.Get("abort_multipart_upload").(*oss.LifecycleAbortMultipartUpload),
				Transitions:          d.Get("transitions").([]oss.LifecycleTransition),
			}

			totalLifeCycle := make([]oss.LifecycleRule, 0)
			totalLifeCycle = append(existedLifeCycle, newLifeCycleRule)

			raw, err := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
				return nil, ossClient.SetBucketLifecycle(bucket, totalLifeCycle)
			})
			if err != nil {
				return WrapErrorf(err, DefaultErrorMsg, d.Id(), "SetBucketLifecycle", AliyunOssGoSdk)
			}
			addDebug("SetBucketLifecycle", raw, requestInfo, map[string]interface{}{
				"bucketName": bucket,
				"rules":      totalLifeCycle,
			})
		}
	}

	return nil
}

func resourceAlicloudOssBucketLifeCycleDelete(d *schema.ResourceData, meta interface{}) error {
	var requestInfo *oss.Client
	client := meta.(*connectivity.AliyunClient)
	bucket := d.Id()

	raw, err := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
		return nil, ossClient.DeleteBucketLifecycle(bucket)
	})
	if err != nil {
		return WrapErrorf(err, DefaultErrorMsg, d.Id(), "SetBucketLifecycle", AliyunOssGoSdk)
	}
	addDebug("DeleteBucketLifecycle", raw, requestInfo, map[string]interface{}{
		"bucketName": bucket,
	})

	return nil
}

func readTheLifeCycleRuleConfiguration(d *schema.ResourceData, meta interface{}) ([]oss.LifecycleRule, error) {
	client := meta.(*connectivity.AliyunClient)
	var requestInfo *oss.Client
	request := map[string]string{"bucketName": d.Id()}
	// Read the lifecycle rule configuration
	raw, err := client.WithOssClient(func(ossClient *oss.Client) (interface{}, error) {
		return ossClient.GetBucketLifecycle(d.Id())
	})
	if err != nil && !ossNotFoundError(err) {
		return nil, WrapErrorf(err, DefaultErrorMsg, d.Id(), "GetBucketLifecycle", AliyunOssGoSdk)
	}
	addDebug("GetBucketLifecycle", raw, requestInfo, request)
	lifeCycles, _ := raw.([]oss.LifecycleRule)

	return lifeCycles, nil
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

func lexpirationHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["date"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["created_before_date"]; ok {
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

func ltransitionsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["created_before_date"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["days"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	if v, ok := m["storage_class"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	return hashcode.String(buf.String())
}

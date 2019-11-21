package alicloud

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/terraform-providers/terraform-provider-alicloud/alicloud/connectivity"
	"strconv"
	"testing"
)

func TestAccAlicloudOssBucketLifecycle(t *testing.T) {
	var v oss.GetBucketInfoResult

	resourceId := "alicloud_oss_bucket.default"
	ra := resourceAttrInit(resourceId, ossBucketBasicMap)

	serviceFunc := func() interface{} {
		return &OssService{testAccProvider.Meta().(*connectivity.AliyunClient)}
	}
	rc := resourceCheckInit(resourceId, &v, serviceFunc)

	rac := resourceAttrCheckInit(rc, ra)

	testAccCheck := rac.resourceAttrMapUpdateSet()
	rand := acctest.RandIntRange(1000000, 9999999)
	name := fmt.Sprintf("tf-testacc-bucket-%d", rand)
	testAccConfig := resourceTestAccConfigFunc(resourceId, name, resourceOssBucketLifeCycleConfigDependence)

	// hashcode for tags
	hashcodeTags1 := strconv.Itoa(tagsHash(map[string]interface{}{
		"key":   "value1",
		"value": "value2",
	}))
	hashcodeTags2 := strconv.Itoa(tagsHash(map[string]interface{}{
		"key":   "value3",
		"value": "value4",
	}))

	// hashcode for expiration
	hashcodeExpiration1 := strconv.Itoa(lexpirationHash(map[string]interface{}{
		"date":                "2020-11-11",
		"created_before_date": "",
		"days":                0,
	}))
	hashcodeExpiration2 := strconv.Itoa(lexpirationHash(map[string]interface{}{
		"date":                "",
		"created_before_date": "2015-11-11",
		"days":                0,
	}))
	hashcodeExpiration3 := strconv.Itoa(lexpirationHash(map[string]interface{}{
		"date":                "",
		"created_before_date": "",
		"days":                365,
	}))

	// hashcode for abort_multipart_upload
	hashcodeAbortMultipartUpload1 := strconv.Itoa(abortMultipartUploadHash(map[string]interface{}{
		"created_before_date": "2015-11-11",
		"days":                0,
	}))
	hashcodeAbortMultipartUpload2 := strconv.Itoa(abortMultipartUploadHash(map[string]interface{}{
		"created_before_date": "",
		"days":                3,
	}))

	// hashcode for transition
	hashcodeTransition1 := strconv.Itoa(ltransitionsHash(map[string]interface{}{
		"days":                3,
		"created_before_date": "",
		"storage_class":       "IA",
	}))
	hashcodeTransition2 := strconv.Itoa(ltransitionsHash(map[string]interface{}{
		"days":                30,
		"created_before_date": "",
		"storage_class":       "Archive",
	}))
	hashcodeTransition3 := strconv.Itoa(ltransitionsHash(map[string]interface{}{
		"days":                0,
		"created_before_date": "2020-11-11",
		"storage_class":       "IA",
	}))
	hashcodeTransition4 := strconv.Itoa(ltransitionsHash(map[string]interface{}{
		"days":                0,
		"created_before_date": "2021-11-11",
		"storage_class":       "Archive",
	}))

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		// module name
		IDRefreshName: resourceId,
		Providers:     testAccProviders,
		CheckDestroy:  rac.checkResourceDestroy(),
		Steps: []resource.TestStep{
			{
				Config: testAccConfig(map[string]interface{}{
					"bucket": name,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"bucket": name,
					}),
				),
			},
			//{
			//	ResourceName:      resourceId,
			//	ImportState:       true,
			//	ImportStateVerify: true,
			//},
			{
				Config: testAccConfig(map[string]interface{}{
					"acl": "public-read-write",
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"acl": "public-read-write",
					}),
				),
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"tags": map[string]string{
						"key1": "value1",
						"Key2": "Value2",
					},
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"tags.%":    "2",
						"tags.key1": "value1",
						"tags.Key2": "Value2",
					}),
				),
			},
			{
				Config: testAccConfig(map[string]interface{}{
					"id":      "rule1",
					"prefix":  "path1/",
					"enabled": "true",
					"tags": map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
					/*// expiration test part
					{
						"id":      "rule2",
						"prefix":  "path2/",
						"enabled": "true",
						"expiration": []map[string]string{
							{
								"date":                "2020-11-11",
								"created_before_date": "",
								"days":                "0",
							},
						},
					},
					{
						"id":      "rule3",
						"prefix":  "path3/",
						"enabled": "true",
						"expiration": []map[string]string{
							{
								"date":                "",
								"created_before_date": "2015-11-11",
								"days":                "0",
							},
						},
					},
					{
						"id":      "rule4",
						"prefix":  "path4/",
						"enabled": "true",
						"expiration": []map[string]string{
							{
								"date":                "",
								"created_before_date": "",
								"days":                "365",
							},
						},
					},

					// abort_multipart_upload test part
					{
						"id":      "rule5",
						"prefix":  "path5/",
						"enabled": "true",
						"abort_multipart_upload": []map[string]string{
							{
								"created_before_date": "2015-11-11",
								"days":                "0",
							},
						},
					},
					{
						"id":      "rule6",
						"prefix":  "path6/",
						"enabled": "true",
						"abort_multipart_upload": []map[string]string{
							{
								"created_before_date": "",
								"days":                "3",
							},
						},
					},

					// transitions test part
					{
						"id":      "rule7",
						"prefix":  "path7/",
						"enabled": "true",
						"transitions": []map[string]interface{}{
							{
								"days":          "3",
								"storage_class": "IA",
							},
							{
								"days":          "30",
								"storage_class": "Archive",
							},
						},
					},
					{
						"id":      "rule8",
						"prefix":  "path8/",
						"enabled": "true",
						"transitions": []map[string]interface{}{
							{
								"created_before_date": "2020-11-11",
								"storage_class":       "IA",
							},
							{
								"created_before_date": "2021-11-11",
								"storage_class":       "Archive",
							},
						},
					},*/
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						// check tags
						"lifecycle_rule.0.id":                               "rule1",
						"lifecycle_rule.0.prefix":                           "path1/",
						"lifecycle_rule.0.enabled":                          "true",
						"lifecycle_rule.0.tags." + hashcodeTags1 + ".key":   "value1",
						"lifecycle_rule.0.tags." + hashcodeTags1 + ".value": "value2",
						"lifecycle_rule.0.tags." + hashcodeTags2 + ".key":   "value3",
						"lifecycle_rule.0.tags." + hashcodeTags2 + ".value": "value4",

						// check expiration
						"lifecycle_rule.1.id":      "rule2",
						"lifecycle_rule.1.prefix":  "path2/",
						"lifecycle_rule.1.enabled": "true",
						"lifecycle_rule.1.expiration." + hashcodeExpiration1 + ".date": "2020-11-11",

						"lifecycle_rule.2.id":      "rule3",
						"lifecycle_rule.2.prefix":  "path3/",
						"lifecycle_rule.2.enabled": "true",
						"lifecycle_rule.2.expiration." + hashcodeExpiration2 + ".created_before_date": "2015-11-11",

						"lifecycle_rule.3.id":      "rule4",
						"lifecycle_rule.3.prefix":  "path4/",
						"lifecycle_rule.3.enabled": "true",
						"lifecycle_rule.3.expiration." + hashcodeExpiration3 + ".days": "365",

						// abort_multipart_upload check
						"lifecycle_rule.4.id":      "rule5",
						"lifecycle_rule.4.prefix":  "path5/",
						"lifecycle_rule.4.enabled": "true",
						"lifecycle_rule.4.abort_multipart_upload." + hashcodeAbortMultipartUpload1 + ".created_before_date": "2015-11-11",

						"lifecycle_rule.5.id":      "rule6",
						"lifecycle_rule.5.prefix":  "path6/",
						"lifecycle_rule.5.enabled": "true",
						"lifecycle_rule.5.abort_multipart_upload." + hashcodeAbortMultipartUpload2 + ".days": "3",

						// transitions check
						"lifecycle_rule.6.id":      "rule7",
						"lifecycle_rule.6.prefix":  "path7/",
						"lifecycle_rule.6.enabled": "true",
						"lifecycle_rule.6.transitions." + hashcodeTransition1 + ".days":          "3",
						"lifecycle_rule.6.transitions." + hashcodeTransition1 + ".storage_class": string(oss.StorageIA),
						"lifecycle_rule.6.transitions." + hashcodeTransition2 + ".days":          "30",
						"lifecycle_rule.6.transitions." + hashcodeTransition2 + ".storage_class": string(oss.StorageArchive),

						"lifecycle_rule.7.id":      "rule8",
						"lifecycle_rule.7.prefix":  "path8/",
						"lifecycle_rule.7.enabled": "true",
						"lifecycle_rule.7.transitions." + hashcodeTransition3 + ".created_before_date": "2020-11-11",
						"lifecycle_rule.7.transitions." + hashcodeTransition3 + ".storage_class":       string(oss.StorageIA),
						"lifecycle_rule.7.transitions." + hashcodeTransition4 + ".created_before_date": "2021-11-11",
						"lifecycle_rule.7.transitions." + hashcodeTransition4 + ".storage_class":       string(oss.StorageArchive)}),
				),
			},

			{
				Config: testAccConfig(map[string]interface{}{
					"acl":            "public-read",
					"tags":           REMOVEKEY,
					"lifecycle_rule": REMOVEKEY,
				}),
				Check: resource.ComposeTestCheckFunc(
					testAccCheck(map[string]string{
						"acl": "public-read",

						"tags.%":    "0",
						"tags.key1": REMOVEKEY,
						"tags.Key2": REMOVEKEY,

						"lifecycle_rule.#":                                  "0",
						"lifecycle_rule.0.id":                               REMOVEKEY,
						"lifecycle_rule.0.prefix":                           REMOVEKEY,
						"lifecycle_rule.0.enabled":                          REMOVEKEY,
						"lifecycle_rule.0.tags." + hashcodeTags1 + ".key":   REMOVEKEY,
						"lifecycle_rule.0.tags." + hashcodeTags1 + ".value": REMOVEKEY,
						"lifecycle_rule.0.tags." + hashcodeTags2 + ".key":   REMOVEKEY,
						"lifecycle_rule.0.tags." + hashcodeTags2 + ".value": REMOVEKEY,

						"lifecycle_rule.1.id":      REMOVEKEY,
						"lifecycle_rule.1.prefix":  REMOVEKEY,
						"lifecycle_rule.1.enabled": REMOVEKEY,
						"lifecycle_rule.1.expiration." + hashcodeExpiration1 + ".date": REMOVEKEY,

						"lifecycle_rule.2.id":      REMOVEKEY,
						"lifecycle_rule.2.prefix":  REMOVEKEY,
						"lifecycle_rule.2.enabled": REMOVEKEY,
						"lifecycle_rule.2.expiration." + hashcodeExpiration2 + ".created_before_date": REMOVEKEY,

						"lifecycle_rule.3.id":      REMOVEKEY,
						"lifecycle_rule.3.prefix":  REMOVEKEY,
						"lifecycle_rule.3.enabled": REMOVEKEY,
						"lifecycle_rule.3.expiration." + hashcodeExpiration3 + ".days": REMOVEKEY,

						"lifecycle_rule.4.id":      REMOVEKEY,
						"lifecycle_rule.4.prefix":  REMOVEKEY,
						"lifecycle_rule.4.enabled": REMOVEKEY,
						"lifecycle_rule.4.abort_multipart_upload." + hashcodeAbortMultipartUpload1 + ".created_before_date": REMOVEKEY,

						"lifecycle_rule.5.id":      REMOVEKEY,
						"lifecycle_rule.5.prefix":  REMOVEKEY,
						"lifecycle_rule.5.enabled": REMOVEKEY,
						"lifecycle_rule.5.abort_multipart_upload." + hashcodeAbortMultipartUpload2 + ".days": REMOVEKEY,

						"lifecycle_rule.6.id":      REMOVEKEY,
						"lifecycle_rule.6.prefix":  REMOVEKEY,
						"lifecycle_rule.6.enabled": REMOVEKEY,
						"lifecycle_rule.6.transitions." + hashcodeTransition1 + ".days":          REMOVEKEY,
						"lifecycle_rule.6.transitions." + hashcodeTransition1 + ".storage_class": REMOVEKEY,
						"lifecycle_rule.6.transitions." + hashcodeTransition2 + ".days":          REMOVEKEY,
						"lifecycle_rule.6.transitions." + hashcodeTransition2 + ".storage_class": REMOVEKEY,

						"lifecycle_rule.7.id":      REMOVEKEY,
						"lifecycle_rule.7.prefix":  REMOVEKEY,
						"lifecycle_rule.7.enabled": REMOVEKEY,
						"lifecycle_rule.7.transitions." + hashcodeTransition3 + ".created_before_date": REMOVEKEY,
						"lifecycle_rule.7.transitions." + hashcodeTransition3 + ".storage_class":       REMOVEKEY,
						"lifecycle_rule.7.transitions." + hashcodeTransition4 + ".created_before_date": REMOVEKEY,
						"lifecycle_rule.7.transitions." + hashcodeTransition4 + ".storage_class":       REMOVEKEY,
					}),
				),
			},
		},
	})
}

func resourceOssBucketLifeCycleConfigDependence(name string) string {
	return fmt.Sprintf(`
resource "alicloud_oss_bucket" "target"{
	bucket = "%s-t"
}
`, name)
}

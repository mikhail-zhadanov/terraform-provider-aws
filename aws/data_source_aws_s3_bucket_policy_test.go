package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceS3BucketPolicy_basic(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-test-bucket")
	//region := testAccGetRegion()
	//hostedZoneID, _ := HostedZoneIDForRegion(region)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceS3BucketPolicyConfig_basic(bucketName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSS3BucketPolicyExists("data.aws_s3_bucket_policy.policy"),
					resource.TestCheckResourceAttrPair("data.aws_s3_bucket_policy.policy", "policy", "aws_s3_bucket_policy.policy", "policy"),
					//resource.TestCheckResourceAttr("data.aws_s3_bucket.bucket", "region", region),
					//testAccCheckS3BucketDomainName("data.aws_s3_bucket.bucket", "bucket_domain_name", bucketName),
					//resource.TestCheckResourceAttr("data.aws_s3_bucket.bucket", "bucket_regional_domain_name", testAccBucketRegionalDomainName(bucketName, region)),
					//resource.TestCheckResourceAttr("data.aws_s3_bucket.bucket", "hosted_zone_id", hostedZoneID),
					//resource.TestCheckNoResourceAttr("data.aws_s3_bucket.bucket", "website_endpoint"),
				),
			},
		},
	})
}

func testAccCheckAWSS3BucketPolicyExists(n string) resource.TestCheckFunc {
	return testAccCheckAWSS3BucketPolicyExistsWithProvider(n, func() *schema.Provider { return testAccProvider })
}

func testAccCheckAWSS3BucketPolicyExistsWithProvider(n string, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fmt.Println("s.RootModule().Resources:", s.RootModule().Resources)
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		provider := providerF()

		conn := provider.Meta().(*AWSClient).s3conn
		_, err := conn.GetBucketPolicy(&s3.GetBucketPolicyInput{
			Bucket: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if isAWSErr(err, s3.ErrCodeNoSuchBucket, "") {
				return fmt.Errorf("s3 bucket not found")
			}
			return err
		}
		return nil

	}
}

func testAccAWSDataSourceS3BucketPolicyConfig_basic(bucketName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "bucket" {
  bucket = "%s"

  tags = {
    TestName = "TestAccAWSS3BucketPolicy_basic"
  }
}

resource "aws_s3_bucket_policy" "bucket" {
  bucket = aws_s3_bucket.bucket.id
  policy = jsonencode({
    Version = "2012-10-17"
    Id      = "MYBUCKETPOLICY"
    Statement = [
      {
        Sid       = "IPAllow"
        Effect    = "Deny"
        Principal = "*"
        Action    = "s3:*"
        Resource = [
          aws_s3_bucket.bucket.arn,
          "${aws_s3_bucket.bucket.arn}/*",
        ]
        Condition = {
          IpAddress = {
            "aws:SourceIp" = "8.8.8.8/32"
          }
        }
      },
    ]
  })
}

data "aws_s3_bucket_policy" "policy" {
  bucket = aws_s3_bucket.bucket.bucket
}

`, bucketName)
}

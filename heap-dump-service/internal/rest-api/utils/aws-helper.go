package utils

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sts"
)

func CheckAWSAccess() error {
	svc := sts.New(session.New())
	input := &sts.GetCallerIdentityInput{}

	_, err := svc.GetCallerIdentity(input)
	if err != nil {
		return errors.New(fmt.Sprintf("Error authenticating to AWS: %s", err.(awserr.Error).Message()))
	}
	return nil
}

func GenerateS3Client(bucketName string) (s3iface.S3API, error) {

	cfg := aws.NewConfig().
		WithEC2MetadataDisableTimeoutOverride(true).
		WithCredentialsChainVerboseErrors(true)

	sess := session.Must(session.NewSession(cfg))
	region, err := s3manager.GetBucketRegion(aws.BackgroundContext(), sess, bucketName, endpoints.EuCentral1RegionID)
	if err != nil {
		return nil, err
	}

	s3Svc := s3.New(sess, &aws.Config{
		Region: aws.String(region),
	})

	return s3Svc, nil
}

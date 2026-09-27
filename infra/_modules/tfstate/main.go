package main

import (
	"context"
	"flag"
	"log"

	"github.com/cloudflare/cloudflare-go"
)

func main() {
	var apiToken, accountID, bucket string
	flag.StringVar(&apiToken, "api-token", "", "Cloudflare account ID")
	flag.StringVar(&accountID, "account-id", "", "Cloudflare account ID")
	flag.StringVar(&bucket, "bucket", "", "Cloudflare R2 bucket name")
	flag.Parse()

	if apiToken == "" || accountID == "" || bucket == "" {
		log.Fatal("--api-token, --account-id and --bucket must be provided")
	}

	resourceContainer := cloudflare.AccountIdentifier(accountID)
	ctx := context.Background()

	api, err := cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		log.Fatalf("failed to create Cloudflare API client: %v", err)
	}

	existingBucket, err := api.GetR2Bucket(ctx, resourceContainer, bucket)
	if err == nil {
		log.Printf("bucket %q already exists: %+v\n", bucket, existingBucket)
		return
	}

	createdBucket, err := api.CreateR2Bucket(
		ctx,
		resourceContainer,
		cloudflare.CreateR2BucketParameters{
			Name: bucket,
		},
	)
	if err != nil {
		log.Fatalf("failed to create bucket: %v", err)
	}

	log.Printf("created bucket: %+v\n", createdBucket)
}

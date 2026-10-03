package state

import (
	"context"
	"errors"
	"fmt"

	cloudflare "github.com/cloudflare/cloudflare-go"
)

func EnsureR2Bucket(ctx context.Context, api *cloudflare.API, account, bucket string) error {
	rc := cloudflare.AccountIdentifier(account)
	if _, err := api.GetR2Bucket(ctx, rc, bucket); err == nil {
		return nil
	} else {
		var missing *cloudflare.NotFoundError
		if !errors.As(err, &missing) {
			return fmt.Errorf("inspect state bucket %q: %w", bucket, err)
		}
	}
	_, err := api.CreateR2Bucket(ctx, rc, cloudflare.CreateR2BucketParameters{Name: bucket})
	if err != nil {
		// A competing provisioner or a lost response can leave the bucket ready.
		if _, checkErr := api.GetR2Bucket(ctx, rc, bucket); checkErr == nil {
			return nil
		}
		return fmt.Errorf("create state bucket %q: %w", bucket, err)
	}
	return nil
}

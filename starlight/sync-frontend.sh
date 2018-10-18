#!/usr/bin/env bash

# If you want to use your own S3 bucket, update 'starlight-client'
# with your bucket name.
cd wallet && npm run build && aws s3 sync public s3://starlight-client --delete --acl public-read

# check-stackdriver-go
Nagios Plugin monitoring GCP stack driver metric.

## Build
Build on linux
```
# go get -u cloud.google.com/go/monitoring/apiv3
# go get -u github.com/jessevdk/go-flags
# go build check_stackdriver.go
```

Build on Windows
```
# GOOS=linux GOARCH=amd64 go build check_stackdriver.go
```

## Usage 
Sample command
```
# ./check_stackdriver -g 'sample-project' -a '/path/to/auth-key.json' \
-m 'storage.googleapis.com/storage/object_count' \
-f 'resource.type = "gcs_bucket" AND resource.labels.bucket_name = "name of gcs bucket"' \
-p 300 -d 240 -e 'LAST' -w 50000 -c 100000
```

Arguments  

option|long option|required|discription
---|---|---|---
-g|--project|true|GCP project id.
-a|--auth|false|GCP authenticate key.
-m|--metric|true|Monitoring metric.
-f|--filter|flase|Filter query.
-d|--delay|false|Shift the acquisition period. (min)
-p|--period|flase|Metric acquisition period. (min)
-e|--evalution|false|Metric evalute type. ("MAX", "LAST", "SUM")
-w|--warning|false|Warning threshold.
-c|--critical|flase|Critical threshold.

# Stack driver metric list
https://cloud.google.com/monitoring/api/metrics?hl=ja

# TODO
[] append evalution type "MEAN".  

# check-stackdriver-go
Nagios Plugin for GCP StackDriver metric monitoring.

## Build

```
go build
```

## Usage 
Sample command
```
# ./check-stackdriver-go -g 'sample-project' -a '/path/to/auth-key.json' \
-m 'storage.googleapis.com/storage/object_count' \
-f 'resource.type = "gcs_bucket" AND resource.labels.bucket_name = "name of gcs bucket"' \
-p 300 -d 240 -e 'LAST' -w 50000 -c 100000
```

Arguments  

option|long option|required|discription
---|---|---|---
-g|--project|true|GCP project id.
-a|--auth|true|GCP authenticate key.
-m|--metric|true|Monitoring metric.
-f|--filter|flase|Filter query.
-d|--delay|false|Shift the acquisition period. (min)
-p|--period|flase|Metric acquisition period. (min)
-e|--evalution|false|Metric evalute type. ("MAX", "LAST", "SUM")
-w|--warning|false|Warning threshold.
-c|--critical|flase|Critical threshold.

## Stack driver metric list
https://pkg.go.dev/google.golang.org/api/monitoring/v3?tab=doc


## Trademark
Nagios, the Nagios logo, and Nagios graphics are the servicemarks, trademarks, or registered trademarks owned by Nagios Enterprises. All other servicemarks and trademarks are the property of their respective owner. 

## License
MIT
# check-stackdriver-go
Nagios Plugin monitoring GCP stack driver metric.

## How to Build
Build on linux
```
# go get -u cloud.google.com/go/monitoring/apiv3
# go get -u github.com/jessevdk/go-flags
# go build check_stack_driver.go
```

## How to use 
Sample command
```
# ./check_stack_driver -g 'sample-project' -a '/path/to/auth-key.json' \
-m 'compute.googleapis.com/instance/cpu/usage_time' \
-p 300 -d 240 -e 'LAST' -w 50000 -c 100000
```

Arguments  

option|long option|required|discription
---|---|---|---
-g|--project|true|GCP project id.
-a|--auth|false|GCP authenticate key.
-m|--metric|true|Monitoring metric.
-d|--delay|false|Shift the acquisition period.
-p|--period|flase|Metric acquisition period.
-e|--evalution|false|Metric evalute type.
-w|--warning|false|Warning threshold.
-c|--critical|flase|Critical threshold.

# Stack driver metric list
https://cloud.google.com/monitoring/api/metrics?hl=ja

# TODO
[] append evalution type "MEAN".
[] append "filter" option.

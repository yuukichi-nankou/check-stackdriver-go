# check-stackdriver-go

## プラグインの使い方 
事前にGCPのSDKを取得して、プラグインをビルドする
```
# go get -u cloud.google.com/go/monitoring/apiv3
# go get -u github.com/jessevdk/go-flags
# go build check_stack_driver.go
```

コマンドの実行

```
# ./check_stack_driver -g 'sample-project' -a '/path/to/auth-key.json' -m 'compute.googleapis.com/instance/cpu/usage_time' \
-p 300 -d 240 -e 'LAST' -w 50000 -c 100000
```


# Stack driver metric list
https://cloud.google.com/monitoring/api/metrics?hl=ja

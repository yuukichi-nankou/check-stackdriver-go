package main

import (
	"fmt"
	"os"

	//"strconv"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	googlepb "github.com/golang/protobuf/ptypes/timestamp"
	flags "github.com/jessevdk/go-flags"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

type Options struct {
	Project   string  `short:"g" long:"project"   required:"true"  description:"GCP project id." `
	Auth      string  `short:"a" long:"auth"      required:"true"  default:"~/gcp_auth_key.json" description:"GCP authenticate key." `
	Metric    string  `short:"m" long:"metric"    required:"true"  description:"Monitoring metric." `
	Filter    string  `short:"f" long:"filter"    required:"false" default:""    description:"Filter query." `
	Delay     int64   `short:"d" long:"delay"     required:"false" default:"4"   description:"Shift the acquisition period." `
	Period    int64   `short:"p" long:"period"    required:"false" default:"5"   description:"Metric acquisition period." `
	Evalution string  `short:"e" long:"evalution" required:"false" default:"MAX" description:"Metric evaluate type." `
	Critical  float64 `short:"c" long:"critical"  required:"false" default:"0.0" description:"Critical threshold." `
	Warning   float64 `short:"w" long:"warning"   required:"false" default:"0.0" description:"Warning threshold." `
	Verbose   []bool  `short:"v" long:"verbose"   required:"false" description:"Verbose option." `
}

func main() {
	// 引数解析処理
	var opts Options
	parser := flags.NewParser(&opts, flags.IgnoreUnknown)
	_, err := parser.Parse()
	if err != nil {
		parser.WriteHelp(os.Stdout)
		output(UNKNOWN, "Missing required arguments.")
	}
	verbose(opts.Verbose, opts)
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", opts.Auth)

	ctx := context.Background()
	c, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		output(UNKNOWN, "GCP SDK Client request failed.")
	}

	var filter string = fmt.Sprintf("metric.type = \"%s\" ", opts.Metric)
	if len(opts.Filter) != 0 {
		filter += fmt.Sprintf("AND %s ", opts.Filter)
	}
	verbose(opts.Verbose, filter)

	unixNow := time.Now().Unix()
	req := &monitoringpb.ListTimeSeriesRequest{
		Name:   "projects/" + opts.Project,
		Filter: filter,
		Interval: &monitoringpb.TimeInterval{
			EndTime: &googlepb.Timestamp{
				Seconds: unixNow - (opts.Delay * 60),
			},
			StartTime: &googlepb.Timestamp{
				Seconds: unixNow - ((opts.Delay + opts.Period) * 60),
			},
		},
	}

	var value float64
	var length int
	it := c.ListTimeSeries(ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			verbose(opts.Verbose, err)
			output(UNKNOWN, "Failed to fetch time series.")
		}
		verbose(opts.Verbose, resp.Metric)
		verbose(opts.Verbose, resp.Resource)
		value = evaluate(opts.Evalution, resp.ValueType.String(), resp.Points)
		verbose(opts.Verbose, value)
		length = len(resp.Points)
	}

	if length == 0 {
		output(UNKNOWN, "Time series is empty.")
	}

	status := OK
	message := ""
	switch {
	case (opts.Critical > 0.0 && value >= opts.Critical):
		status = CRITICAL
		message = fmt.Sprintf("%s value: %d over %d", opts.Evalution, int(value), int(opts.Critical))
	case (opts.Warning > 0.0 && value >= opts.Warning):
		status = WARNING
		message = fmt.Sprintf("%s value: %d over %d", opts.Evalution, int(value), int(opts.Warning))
	default:
		status = OK
		message = fmt.Sprintf("%s value: %d ", opts.Evalution, int(value))
	}
	paformace := fmt.Sprintf("|value=%f;%d;%d", value, int(opts.Warning), int(opts.Critical))
	output(status, message+paformace)
}

const (
	OK = iota
	WARNING
	CRITICAL
	UNKNOWN
)

func output(status int, message string) {
	switch status {
	case OK:
		message = "OK - " + message
	case WARNING:
		message = "WARNING - " + message
	case CRITICAL:
		message = "CRITICAL - " + message
	case UNKNOWN:
		message = "UNKNOWN - " + message
	default:
		message = "UNKNOWN - " + message
	}
	fmt.Println(message)
	os.Exit(status)
}

func evaluate(evaluateType string, valueType string, points []*monitoringpb.Point) float64 {
	var ret float64
	switch evaluateType {
	case "LAST":
		ret = getFloatValue(valueType, points[0].GetValue())
	case "SUM":
		for _, point := range points {
			ret += getFloatValue(valueType, point.GetValue())
		}
	case "MAX":
		var current float64
		for _, point := range points {
			current = getFloatValue(valueType, point.GetValue())
			if current < ret {
				continue
			}
			ret = current
		}
	}
	return ret
}

func getFloatValue(valueType string, typedValue *monitoringpb.TypedValue) float64 {
	var ret float64
	switch valueType {
	case "INT64":
		ret = float64(typedValue.GetInt64Value())
	case "DOUBLE":
		ret = typedValue.GetDoubleValue()
	case "DISTRIBUTION":
		ret = typedValue.GetDistributionValue().GetMean()
	default:
		// Expected "BOOL" "STRING" "MONEY", these cases are unsupported.
	}
	return ret
}

func verbose(flag []bool, value interface{}) {
	if len(flag) == 0 {
		return
	}
	if flag[0] {
		fmt.Println(value)
	}
}

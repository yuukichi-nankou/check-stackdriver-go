package main

import (
	"fmt"
	"os"
	"sort"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	googlepb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/jessevdk/go-flags"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

const (
	OK = iota
	WARNING
	CRITICAL
	UNKNOWN
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

type CheckResult struct {
	Status            int
	ResourceID        string
	ResourceFullName  string
	ThresholdWarning  float64
	ThresholdCritical float64
	Metric            float64
	Message           string
}

func main() {
	var opts Options
	parser := flags.NewParser(&opts, flags.IgnoreUnknown)
	_, err := parser.Parse()
	if err != nil {
		parser.WriteHelp(os.Stdout)
		fmt.Printf("UNKNOWN: Missing required arguments. \n")
		os.Exit(UNKNOWN)
	}
	verbose(opts.Verbose, opts)
	err = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", opts.Auth)
	if err != nil {
		fmt.Printf("UNKNOWN: Missing required arguments. \n")
		os.Exit(UNKNOWN)
	}

	ctx := context.Background()
	c, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		fmt.Printf("UNKNOWN: GCP SDK Client request failed. \n")
		os.Exit(UNKNOWN)
	}

	var filter = fmt.Sprintf("metric.type = \"%s\" ", opts.Metric)
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

	status := OK
	resourceType := ""
	var checkResults []CheckResult

	it := c.ListTimeSeries(ctx, req)
	for {
		resp, err := it.Next()

		if err == iterator.Done {
			break
		}

		if err != nil {
			verbose(opts.Verbose, err)
			fmt.Printf("UNKNOWN: Failed to fetch time series. \n")
			os.Exit(UNKNOWN)
		}

		var value CheckResult

		value.ThresholdWarning = opts.Warning
		value.ThresholdCritical = opts.Critical
		resourceType = resp.Resource.Type

		verbose(opts.Verbose, resp.Metric)
		verbose(opts.Verbose, resp.Resource)

		length := len(resp.Points)

		if length == 0 {
			status = UNKNOWN
			value.Status = UNKNOWN
			value.Message = "Time series is empty."
			continue
		}

		value.Metric = evaluate(opts.Evalution, resp.ValueType.String(), resp.Points)
		verbose(opts.Verbose, value.Metric)

		labelsMap := resp.Resource.GetLabels()

		value.ResourceID = labelsMap["instance_id"]
		value.ResourceFullName = "Project=" + labelsMap["project_id"] + " Zone=" + labelsMap["zone"] + " ID=" + labelsMap["instance_id"]

		if opts.Warning > 0.0 && value.Metric < opts.Warning {
			value.Status = OK
			value.Message = metric2string(value, opts)
		}

		if opts.Warning > 0.0 && value.Metric >= opts.Warning {
			if status <= WARNING {
				status = WARNING
			}
			value.Status = WARNING
			value.Message = metric2string(value, opts)
		}

		if opts.Critical > 0.0 && value.Metric >= opts.Critical {
			status = CRITICAL
			value.Status = CRITICAL
			value.Message = metric2string(value, opts)
		}

		checkResults = append(checkResults, value)
	}

	if status == OK {
		performance := getPerformanceData(checkResults)
		fmt.Printf("OK: All %d %s metrics less then tresholds %s \n", len(checkResults), resourceType, performance)
	} else {
		output(checkResults)
	}

	os.Exit(status)
}

func getPerformanceData(data []CheckResult) string {
	performance := "|"
	for _, value := range data {
		if value.Status == UNKNOWN {
			performance += fmt.Sprintf("'%s'=U;%d;%d;0;0 ", value.ResourceID, int(value.ThresholdWarning), int(value.ThresholdCritical))
		} else {
			performance += fmt.Sprintf("'%s'=%f;%d;%d;0;0 ", value.ResourceID, value.Metric, int(value.ThresholdWarning), int(value.ThresholdCritical))
		}
	}
	return performance
}

func metric2string(data CheckResult, opts Options) string {
	statusStr := ""
	switch data.Status {
	case OK:
		statusStr = "OK"
	case WARNING:
		statusStr = "WARNING"
	case CRITICAL:
		statusStr = "CRITICAL"
	case UNKNOWN:
		statusStr = "UNKNOWN"
	default:
		statusStr = "UNKNOWN"
	}
	return fmt.Sprintf("%s: %s [warn=%d:crit=%d:datapoints=%d]\n", statusStr, data.ResourceFullName, int(opts.Warning), int(opts.Critical), int(data.Metric))
}

func output(results []CheckResult) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].Status > results[j].Status
	})

	message := ""
	for _, value := range results {
		message += value.Message
	}

	performance := getPerformanceData(results)

	fmt.Println(message + performance)
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

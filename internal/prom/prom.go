package prom

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const baseURL = "http://localhost:9090/api/v1"

type Sample struct {
	T float64 `json:"t"`
	V float64 `json:"v"`
}

type RangeOptions struct {
	Vhost string
	Queue string
	Since time.Duration
	Step  time.Duration
}

func QueryRange(opts RangeOptions) ([]Sample, error) {
	query := fmt.Sprintf(`rabbitmq_detailed_queue_messages{vhost=%q,queue=%q}`,
		opts.Vhost, opts.Queue)

	now := time.Now()
	start := now.Add(-opts.Since)

	params := url.Values{}
	params.Set("query", query)
	params.Set("start", strconv.FormatInt(start.Unix(), 10))
	params.Set("end", strconv.FormatInt(now.Unix(), 10))
	params.Set("step", strconv.FormatInt(int64(opts.Step.Seconds()), 10))

	resp, err := http.Get(baseURL + "/query_range?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("prometheus unreachable: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			Result []struct {
				Values [][2]any `json:"values"`
			} `json:"result"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result.Data.Result) == 0 {
		return nil, nil
	}

	raw := result.Data.Result[0].Values
	samples := make([]Sample, 0, len(raw))
	for _, v := range raw {
		ts, _ := v[0].(float64)
		valStr, _ := v[1].(string)
		val, _ := strconv.ParseFloat(valStr, 64)
		samples = append(samples, Sample{T: ts, V: val})
	}
	return samples, nil
}

// StepFor returns a sensible Prometheus step for a given time range.
func StepFor(since time.Duration) time.Duration {
	switch {
	case since <= 10*time.Minute:
		return 15 * time.Second
	case since <= time.Hour:
		return 30 * time.Second
	case since <= 6*time.Hour:
		return 2 * time.Minute
	case since <= 24*time.Hour:
		return 5 * time.Minute
	default:
		return 30 * time.Minute
	}
}

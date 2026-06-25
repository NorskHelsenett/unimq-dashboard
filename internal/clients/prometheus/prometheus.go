package prometheus

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rest"
	"github.com/sisneve/rabbitmq-dashboard/internal/clients/rest/httpauthproviders"
	"github.com/sisneve/rabbitmq-dashboard/internal/models"
)

type PromClient struct {
	RestClient *rest.RestClient
	apiURL     string
	port       int
	apiVersion string
}

func NewPromClient(baseURL, apiVersion, username, password string, port int) (*PromClient, error) {
	url := fmt.Sprintf("%v:%d/api/%v", baseURL, port, apiVersion)
	restclient, err := rest.NewRestClient(url,
		rest.WithAuthProvider(httpauthproviders.NewBasicAuthProvider(username, password)),
	)

	if err != nil {
		return nil, err
	}
	return &PromClient{
		RestClient: restclient,
		apiURL:     url,
		port:       port,
		apiVersion: apiVersion,
	}, nil
}

func (pc *PromClient) QueryRange(opts models.RangeOptions) ([]models.Sample, error) {
	query := fmt.Sprintf(`rabbitmq_detailed_queue_messages{vhost=%q,queue=%q}`,
		opts.Vhost, opts.Queue)

	now := time.Now()
	start := now.Add(-opts.Since)

	params := url.Values{}
	params.Set("query", query)
	params.Set("start", strconv.FormatInt(start.Unix(), 10))
	params.Set("end", strconv.FormatInt(now.Unix(), 10))
	params.Set("step", strconv.FormatInt(int64(opts.Step.Seconds()), 10))

	resp, err := http.Get(pc.apiURL + "/query_range?" + params.Encode())
	if err != nil {
		return nil, fmt.Errorf("prometheus unreachable: %w", err)
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			slog.Error("failed to close cursor", "runtime", time.Since(start), "error", err)
		}
	}()

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
	samples := make([]models.Sample, 0, len(raw))
	for _, v := range raw {
		ts, _ := v[0].(float64)
		valStr, _ := v[1].(string)
		val, _ := strconv.ParseFloat(valStr, 64)
		samples = append(samples, models.Sample{T: ts, V: val})
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

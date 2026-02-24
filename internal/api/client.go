package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	runBase   = "https://fal.run"
	queueBase = "https://queue.fal.run"
	apiBase   = "https://api.fal.ai/v1"
)

// Client is an authenticated fal.ai API client.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a new authenticated Client.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// authHeader returns the Authorization header value.
func (c *Client) authHeader() string {
	return "Key " + c.apiKey
}

// doRequest executes an HTTP request and returns the body bytes.
func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("Authorization", c.authHeader())
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		// Try to parse fal error format
		var falErr FalError
		if json.Unmarshal(body, &falErr) == nil && falErr.Detail != "" {
			return nil, &falErr
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// postJSON makes a POST request with a JSON body to the given full URL.
func (c *Client) postJSON(fullURL string, payload any) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encoding request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fullURL, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.doRequest(req)
}

// get makes a GET request to the given full URL with optional query params.
func (c *Client) get(fullURL string, params url.Values) ([]byte, error) {
	if len(params) > 0 {
		u, err := url.Parse(fullURL)
		if err != nil {
			return nil, err
		}
		q := u.Query()
		for k, vs := range params {
			for _, v := range vs {
				q.Add(k, v)
			}
		}
		u.RawQuery = q.Encode()
		fullURL = u.String()
	}

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}

	return c.doRequest(req)
}

// put makes a PUT request to the given full URL (used for cancel).
func (c *Client) put(fullURL string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPut, fullURL, nil)
	if err != nil {
		return nil, err
	}

	return c.doRequest(req)
}

// ---- Sync run ----

// RunSync submits a synchronous request to a model and returns the raw response.
// modelID is e.g. "fal-ai/nano-banana-pro" or "fal-ai/nano-banana-pro/edit".
func (c *Client) RunSync(modelID string, payload any) ([]byte, error) {
	endpoint := runBase + "/" + strings.TrimPrefix(modelID, "/")
	return c.postJSON(endpoint, payload)
}

// ---- Queue ----

// QueueSubmit submits a request to the queue and returns queue metadata.
func (c *Client) QueueSubmit(modelID string, payload any) (*QueueSubmitResponse, error) {
	endpoint := queueBase + "/" + strings.TrimPrefix(modelID, "/")
	body, err := c.postJSON(endpoint, payload)
	if err != nil {
		return nil, err
	}

	var resp QueueSubmitResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing queue response: %w", err)
	}
	return &resp, nil
}

// QueueStatus polls the status of a queued request.
// modelID is required to build the status URL.
func (c *Client) QueueStatus(modelID, requestID string, withLogs bool) (*QueueStatus, error) {
	endpoint := fmt.Sprintf("%s/%s/requests/%s/status",
		queueBase, strings.TrimPrefix(modelID, "/"), requestID)

	params := url.Values{}
	if withLogs {
		params.Set("logs", "1")
	}

	body, err := c.get(endpoint, params)
	if err != nil {
		return nil, err
	}

	var status QueueStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("parsing status: %w", err)
	}
	return &status, nil
}

// QueueResult retrieves the completed result for a request.
func (c *Client) QueueResult(modelID, requestID string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/%s/requests/%s",
		queueBase, strings.TrimPrefix(modelID, "/"), requestID)
	return c.get(endpoint, nil)
}

// QueueCancel cancels a queued request.
func (c *Client) QueueCancel(modelID, requestID string) error {
	endpoint := fmt.Sprintf("%s/%s/requests/%s/cancel",
		queueBase, strings.TrimPrefix(modelID, "/"), requestID)
	_, err := c.put(endpoint)
	return err
}

// ---- Queue + poll helper ----

// RunQueued submits to queue, polls until done, then returns the result.
// progress is called on each status poll if non-nil.
func (c *Client) RunQueued(modelID string, payload any, progress func(status *QueueStatus)) ([]byte, error) {
	sub, err := c.QueueSubmit(modelID, payload)
	if err != nil {
		return nil, err
	}

	// Poll with backoff: 1s, 2s, 4s, 8s, then every 10s
	delay := time.Second
	maxDelay := 10 * time.Second

	for {
		time.Sleep(delay)
		if delay < maxDelay {
			delay *= 2
			if delay > maxDelay {
				delay = maxDelay
			}
		}

		status, err := c.QueueStatus(modelID, sub.RequestID, progress != nil)
		if err != nil {
			return nil, err
		}

		if progress != nil {
			progress(status)
		}

		if status.Status == "COMPLETED" {
			return c.QueueResult(modelID, sub.RequestID)
		}
	}
}

// ---- Platform API: Models ----

// ListModels fetches models from the catalog with optional filters.
func (c *Client) ListModels(q, category, cursor string, limit int) (*ModelsResponse, error) {
	params := url.Values{}
	if q != "" {
		params.Set("q", q)
	}
	if category != "" {
		params.Set("category", category)
	}
	if cursor != "" {
		params.Set("cursor", cursor)
	}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}

	body, err := c.get(apiBase+"/models", params)
	if err != nil {
		return nil, err
	}

	var resp ModelsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing models: %w", err)
	}
	return &resp, nil
}

// GetModelPricing fetches pricing for one or more model endpoint IDs.
func (c *Client) GetModelPricing(endpointIDs []string) (*PricingResponse, error) {
	params := url.Values{}
	for _, id := range endpointIDs {
		params.Add("endpoint_id", id)
	}

	body, err := c.get(apiBase+"/models/pricing", params)
	if err != nil {
		return nil, err
	}

	var resp PricingResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parsing pricing: %w", err)
	}
	return &resp, nil
}

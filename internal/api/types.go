package api

import "encoding/json"

// ---- Queue / Run types ----

// QueueSubmitResponse is returned when submitting a request to the queue.
type QueueSubmitResponse struct {
	RequestID   string `json:"request_id"`
	ResponseURL string `json:"response_url"`
	StatusURL   string `json:"status_url"`
	CancelURL   string `json:"cancel_url"`
}

// QueueStatus represents the current status of a queued request.
type QueueStatus struct {
	Status        string     `json:"status"` // IN_QUEUE, IN_PROGRESS, COMPLETED
	QueuePosition *int       `json:"queue_position,omitempty"`
	Logs          []LogEntry `json:"logs,omitempty"`
}

// LogEntry is a single log line from a running model.
type LogEntry struct {
	Message   string `json:"message"`
	Level     string `json:"level"`
	Source    string `json:"source"`
	Timestamp string `json:"timestamp"`
}

// ---- Model catalog types ----

// ModelMetadata holds display info for a model.
type ModelMetadata struct {
	DisplayName  string   `json:"display_name"`
	Category     string   `json:"category"`
	Description  string   `json:"description"`
	Status       string   `json:"status"`
	Tags         []string `json:"tags"`
	UpdatedAt    string   `json:"updated_at"`
	ThumbnailURL string   `json:"thumbnail_url"`
	ModelURL     string   `json:"model_url"`
}

// Model represents a single model from the catalog.
type Model struct {
	EndpointID string        `json:"endpoint_id"`
	Metadata   ModelMetadata `json:"metadata"`
}

// ModelsResponse is the paginated list of models.
type ModelsResponse struct {
	Models     []Model `json:"models"`
	NextCursor string  `json:"next_cursor"`
	HasMore    bool    `json:"has_more"`
}

// ---- Pricing types ----

// ModelPrice is the price entry for a single model.
type ModelPrice struct {
	EndpointID string  `json:"endpoint_id"`
	UnitPrice  float64 `json:"unit_price"`
	Unit       string  `json:"unit"`
	Currency   string  `json:"currency"`
}

// PricingResponse is the response from the pricing endpoint.
type PricingResponse struct {
	Prices     []ModelPrice `json:"prices"`
	NextCursor string       `json:"next_cursor"`
	HasMore    bool         `json:"has_more"`
}

// ---- Image generation types (nano-banana-pro) ----

// ImageFile represents a single generated image.
type ImageFile struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
}

// GenerateResponse is the output from image generation models.
type GenerateResponse struct {
	Images      []ImageFile     `json:"images"`
	Description string          `json:"description"`
	Seed        uint64          `json:"seed"`
	Timings     json.RawMessage `json:"timings"`
}

// ---- File upload types ----

// FileUploadResponse is returned after uploading a local file to fal.ai storage.
type FileUploadResponse struct {
	URL         string `json:"url"`
	ContentType string `json:"content_type"`
	FileName    string `json:"file_name"`
	FileSize    int64  `json:"file_size"`
}

// ---- fal API error ----

// FalError represents an error returned by the fal.ai API.
type FalError struct {
	Detail string `json:"detail"`
	Status int    `json:"status"`
}

func (e *FalError) Error() string {
	if e.Detail != "" {
		return e.Detail
	}
	return "fal API error"
}

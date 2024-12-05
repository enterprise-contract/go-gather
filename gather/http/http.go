package http

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/enterprise-contract/go-gather/gather"
	"github.com/enterprise-contract/go-gather/metadata"
)

type HTTPGatherer struct {
	Client *http.Client
}

func NewHTTPGatherer() *HTTPGatherer {
	return &HTTPGatherer{
		Client: &http.Client{Timeout: 30 * time.Second},
	}
}

type HTTPMetadata struct {
	URI          string
	Path         string
	ResponseCode int
	Size         int64
	Timestamp    string
}

func (h *HTTPMetadata) Get() interface{} {
	return h
}

func (h *HTTPGatherer) Gather(ctx context.Context, src, dst string) (metadata.Metadata, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	req, err := http.NewRequestWithContext(ctx, "GET", src, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := h.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download from URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 response code: %d", resp.StatusCode)
	}

	outFile, err := os.Create(dst)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer outFile.Close()

	bytesWritten, err := io.Copy(outFile, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to write to destination file: %w", err)
	}

	return &HTTPMetadata{
		URI:          src,
		Path:         dst,
		ResponseCode: resp.StatusCode,
		Size:         bytesWritten,
		Timestamp:    time.Now().Format(time.RFC3339)}, nil
}

func (h *HTTPGatherer) Matcher(uri string) bool {
	return strings.HasPrefix(uri, "http://") || strings.HasPrefix(uri, "https://")
}

func init() {
	gather.RegisterGatherer(&HTTPGatherer{})
}

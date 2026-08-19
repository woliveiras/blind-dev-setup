package artifact

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/woliveiras/blind-dev-setup/internal/manifest"
)

func TestFetchDownloadsAndReusesVerifiedCache(t *testing.T) {
	content := []byte("verified artifact")
	requests := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests++
		_, _ = writer.Write(content)
	}))
	defer server.Close()

	sum := sha256.Sum256(content)
	component := manifest.Component{
		ID:            "editor",
		Version:       "1.2.3",
		URL:           server.URL + "/editor.zip",
		SHA256:        fmt.Sprintf("%x", sum),
		DownloadBytes: int64(len(content)),
	}
	fetcher := Fetcher{Client: server.Client()}
	cache := t.TempDir()

	first, err := fetcher.Fetch(context.Background(), component, cache)
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	second, err := fetcher.Fetch(context.Background(), component, cache)
	if err != nil {
		t.Fatalf("Fetch() cached error = %v", err)
	}
	if first != second || requests != 1 {
		t.Fatalf("paths = %q, %q; requests = %d", first, second, requests)
	}
}

func TestFetchRejectsChecksumMismatchAndRemovesPartialFile(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = io.WriteString(writer, "tampered")
	}))
	defer server.Close()

	component := manifest.Component{
		ID:            "editor",
		Version:       "1.2.3",
		URL:           server.URL,
		SHA256:        strings.Repeat("a", 64),
		DownloadBytes: int64(len("tampered")),
	}
	cache := t.TempDir()
	_, err := (Fetcher{Client: server.Client()}).Fetch(context.Background(), component, cache)
	if err == nil || !strings.Contains(err.Error(), "SHA-256") {
		t.Fatalf("Fetch() error = %v", err)
	}
	entries, readErr := os.ReadDir(cache)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("cache contains partial files: %v", entries)
	}
}

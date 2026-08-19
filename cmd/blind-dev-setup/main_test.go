package main

import (
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestDownloadClientHasBoundedTimeoutAndSecureRedirects(t *testing.T) {
	client := downloadClient()
	if client.Timeout != 30*time.Minute {
		t.Fatalf("Timeout = %s", client.Timeout)
	}
	secureRequest := &http.Request{URL: &url.URL{Scheme: "https", Host: "example.test"}}
	if err := client.CheckRedirect(secureRequest, nil); err != nil {
		t.Fatalf("HTTPS redirect error = %v", err)
	}
	insecureRequest := &http.Request{URL: &url.URL{Scheme: "http", Host: "example.test"}}
	if err := client.CheckRedirect(insecureRequest, nil); err == nil {
		t.Fatal("HTTP redirect was accepted")
	}
}

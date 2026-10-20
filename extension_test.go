package tiktoktts

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestExtensionProtocolUsesProvider(t *testing.T) {
	calls := 0
	client := &http.Client{Transport: fakeTransport(func(request *http.Request) (*http.Response, error) {
		calls++
		if request.URL.String() != AlternateEndpoint+"/media/api/text/speech/invoke/" || request.Header.Get("Cookie") != "sessionid="+testSession {
			t.Fatal("wrong provider configuration")
		}
		if err := request.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if request.Form.Get("text_speaker") != "en_us_001" || len(request.Form.Get("req_text")) > 300 {
			t.Fatal("invalid speech request")
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"status_code":0,"data":{"v_str":"YXVkaW8="}}`))}, nil
	})}
	request := Invocation{Protocol: 1, Operation: "speech", Configuration: map[string]string{"endpoint": AlternateEndpoint, "session_token": testSession}, Text: strings.Repeat("👻", 100), Voice: "en_us_001"}
	input, _ := json.Marshal(request)
	var output bytes.Buffer
	if err := Serve(context.Background(), client, bytes.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	var response Result
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Protocol != 1 || response.Error != "" || string(response.Audio) != "audioaudio" || calls != 2 {
		t.Fatalf("response: %+v calls=%d", response, calls)
	}
	if strings.Contains(output.String(), testSession) {
		t.Fatal("credential leaked")
	}
}

func TestExtensionProtocolErrors(t *testing.T) {
	for status, expected := range map[int]string{401: "credentials_rejected", 403: "credentials_rejected", 429: "busy", 500: "provider_failed"} {
		client := &http.Client{Transport: fakeTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(testSession))}, nil
		})}
		input, _ := json.Marshal(Invocation{1, "speech", map[string]string{"endpoint": PrimaryEndpoint, "session_token": testSession}, "hello", "en_us_001"})
		var output bytes.Buffer
		if err := Serve(context.Background(), client, bytes.NewReader(input), &output); err != nil {
			t.Fatal(err)
		}
		var response Result
		if json.Unmarshal(output.Bytes(), &response) != nil || response.Error != expected || len(response.Audio) != 0 || strings.Contains(output.String(), testSession) {
			t.Fatalf("status %d: %s", status, &output)
		}
	}
	for _, input := range []string{`{}`, `{"protocol":2,"operation":"speech"}`, `{"protocol":1,"operation":"unknown"}`, `{} {}`, strings.Repeat("x", 256*1024+1)} {
		var output bytes.Buffer
		if err := Serve(context.Background(), nil, strings.NewReader(input), &output); err != nil {
			t.Fatal(err)
		}
		var response Result
		if json.Unmarshal(output.Bytes(), &response) != nil || response.Error != "invalid_request" {
			t.Fatal(output.String())
		}
	}
}

func TestPackageManifest(t *testing.T) {
	data, err := os.ReadFile("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Schema       int    `json:"schema"`
		Runtime      string `json:"runtime"`
		Entrypoint   string `json:"entrypoint"`
		DefaultVoice string `json:"default_voice"`
		Voices       []struct {
			ID string `json:"id"`
		} `json:"voices"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Schema != 1 || manifest.Runtime != "stdio-v1" || manifest.Entrypoint != "extension" || len(manifest.Voices) != 42 {
		t.Fatal("invalid package metadata")
	}
	found := false
	for _, voice := range manifest.Voices {
		found = found || voice.ID == manifest.DefaultVoice
	}
	if !found {
		t.Fatal("default voice missing from menu")
	}
}

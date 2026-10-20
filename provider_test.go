package tiktoktts

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
)

type fakeTransport func(*http.Request) (*http.Response, error)

func (transport fakeTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

const testSession = "abcdef0123456789abcdef0123456789"

func TestMobileSpeechAndByteSafeChunking(t *testing.T) {
	calls := 0
	client := &http.Client{Transport: fakeTransport(func(request *http.Request) (*http.Response, error) {
		calls++
		if request.Header.Get("Cookie") != "sessionid="+testSession || !strings.Contains(request.UserAgent(), "Android") {
			t.Fatal("missing mobile authentication")
		}
		if err := request.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if len(request.Form.Get("req_text")) > 300 || request.Form.Get("aid") != "1233" {
			t.Fatal("invalid request form")
		}
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`{"status_code":0,"data":{"v_str":"YXVkaW8="}}`))}, nil
	})}
	audio, err := Synthesize(context.Background(), client, PrimaryEndpoint, testSession, Request{Text: strings.Repeat("👻", 100), Voice: "en_us_001"})
	if err != nil || calls != 2 || string(audio) != "audioaudio" {
		t.Fatalf("speech: %q %v calls=%d", audio, err, calls)
	}
}

func TestRejectForeignHostsAndOversizedText(t *testing.T) {
	for _, prompt := range []struct{ endpoint, text string }{{"https://attacker.test", "hello"}, {PrimaryEndpoint, strings.Repeat("x", 501)}} {
		if _, err := Synthesize(context.Background(), nil, prompt.endpoint, testSession, Request{Text: prompt.text, Voice: "en_us_001"}); !errors.Is(err, ErrInvalid) {
			t.Fatal("invalid request accepted", err)
		}
	}
}

func TestRedirectDoesNotForwardSession(t *testing.T) {
	calls := 0
	client := &http.Client{Transport: fakeTransport(func(request *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 302, Header: http.Header{"Location": {"https://attacker.test"}}, Body: io.NopCloser(strings.NewReader(testSession))}, nil
	})}
	_, err := Synthesize(context.Background(), client, PrimaryEndpoint, testSession, Request{Text: "hello", Voice: "en_us_001"})
	if !errors.Is(err, ErrUpstream) || calls != 1 || strings.Contains(err.Error(), testSession) {
		t.Fatal("unsafe redirect or error")
	}
}

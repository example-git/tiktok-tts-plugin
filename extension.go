package tiktoktts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type Invocation struct {
	Protocol      int               `json:"protocol"`
	Operation     string            `json:"operation"`
	Configuration map[string]string `json:"configuration"`
	Text          string            `json:"text"`
	Voice         string            `json:"voice"`
}

type Result struct {
	Protocol int    `json:"protocol"`
	Audio    []byte `json:"audio,omitempty"`
	Error    string `json:"error,omitempty"`
}

func Serve(ctx context.Context, client *http.Client, input io.Reader, output io.Writer) error {
	response := Result{Protocol: 1}
	data, err := io.ReadAll(io.LimitReader(input, 256*1024+1))
	var request Invocation
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err != nil || len(data) > 256*1024 || decoder.Decode(&request) != nil || decoder.Decode(new(any)) != io.EOF || request.Protocol != 1 || request.Operation != "speech" {
		response.Error = "invalid_request"
	} else {
		response.Audio, err = Synthesize(ctx, client, request.Configuration["endpoint"], request.Configuration["session_token"], Request{Text: request.Text, Voice: request.Voice})
		switch {
		case errors.Is(err, ErrInvalid):
			response.Error = "invalid_request"
		case errors.Is(err, ErrCredentials):
			response.Error = "credentials_rejected"
		case errors.Is(err, ErrBusy):
			response.Error = "busy"
		case err != nil:
			response.Error = "provider_failed"
		}
	}
	return json.NewEncoder(output).Encode(response)
}

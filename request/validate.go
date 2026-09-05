package request

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/lsgser/gofreight/validation"
	"github.com/lsgser/gofreight/vine"
)

// All returns merged request input from form values and query string.
func All(r *http.Request) (map[string]string, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	data := make(map[string]string)
	for k, v := range r.Form {
		if len(v) > 0 {
			data[k] = v[0]
		}
	}
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			if _, exists := data[k]; !exists {
				data[k] = v[0]
			}
		}
	}
	return data, nil
}

// ValidateUsing validates request input with a Vine schema.
func ValidateUsing(r *http.Request, schema *vine.ObjectSchema) (map[string]string, validation.Errors, error) {
	data, err := All(r)
	if err != nil {
		return nil, nil, err
	}
	if wantsJSON(r) && len(data) == 0 {
		data, err = jsonBody(r)
		if err != nil {
			return nil, nil, err
		}
	}
	_, errs := schema.Validate(data)
	return data, errs, nil
}

func jsonBody(r *http.Request) (map[string]string, error) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	var anyMap map[string]any
	if err := json.Unmarshal(raw, &anyMap); err != nil {
		return nil, err
	}
	out := make(map[string]string, len(anyMap))
	for k, v := range anyMap {
		out[k] = stringify(v)
	}
	return out, nil
}

func stringify(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case bool:
		if t {
			return "1"
		}
		return "0"
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		return fmt.Sprint(v)
	}
}

func wantsJSON(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	ct := r.Header.Get("Content-Type")
	return strings.Contains(accept, "application/json") ||
		strings.Contains(ct, "application/json")
}

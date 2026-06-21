package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func decodeOptionalJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	if r.Body == nil {
		return nil
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return fmt.Errorf("decode json: %w", err)
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must contain a single JSON object")
	}

	return nil
}

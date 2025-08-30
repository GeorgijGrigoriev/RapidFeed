package http

import "encoding/json"

func jsonError(code int, err string) string {
	body := map[string]any{"status": code, "error": err}

	marshaledBody, marshalError := json.Marshal(body)
	if marshalError != nil {
		panic(marshalError)
	}

	return string(marshaledBody)
}

func jsonResponse(code int, content any) string {
	body := map[string]any{"status": code, "content": content}

	marshaledBody, marshalError := json.Marshal(body)
	if marshalError != nil {
		panic(marshalError)
	}

	return string(marshaledBody)
}

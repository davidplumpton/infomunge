package formats

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"sort"
	"strings"

	unifiederrors "infomunge/internal/errors"
)

const multipartBoundary = "infomunge-boundary"

func init() {
	RegisterReader("multipart/form-data", readMultipart)
	RegisterWriter("multipart/form-data", formatMultipart)
	RegisterExtension(".multipart", "multipart/form-data")
	RegisterExtension(".formdata", "multipart/form-data")
}

func readMultipart(content string) (interface{}, error) {
	boundary, err := detectMultipartBoundary(content)
	if err != nil {
		return nil, err
	}

	reader := multipart.NewReader(strings.NewReader(content), boundary)
	result := make(Object)

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, unifiederrors.WrapValidationf(err, "multipart parse error: %v", err)
		}

		name := part.FormName()
		if name == "" {
			continue
		}

		body, err := io.ReadAll(part)
		if err != nil {
			return nil, unifiederrors.WrapValidationf(err, "multipart read error: %v", err)
		}

		value := multipartPartValue(part, string(body))
		addMultipartValue(result, name, value)
	}

	return result, nil
}

func detectMultipartBoundary(content string) (string, error) {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, "--") {
			return "", unifiederrors.ValidationError("multipart input must start with boundary delimiter")
		}
		boundary := strings.TrimPrefix(trimmed, "--")
		boundary = strings.TrimSuffix(boundary, "--")
		boundary = strings.TrimSpace(boundary)
		if boundary == "" {
			return "", unifiederrors.ValidationError("multipart input boundary cannot be empty")
		}
		return boundary, nil
	}

	return "", unifiederrors.ValidationError("multipart input is empty")
}

func multipartPartValue(part *multipart.Part, content string) interface{} {
	filename := part.FileName()
	contentType := strings.TrimSpace(part.Header.Get("Content-Type"))

	if filename == "" && (contentType == "" || strings.EqualFold(contentType, "text/plain")) {
		return content
	}

	result := Object{
		"content": content,
	}
	if filename != "" {
		result["filename"] = filename
	}
	if contentType != "" {
		result["contentType"] = contentType
	}
	return result
}

func addMultipartValue(target Object, key string, value interface{}) {
	current, exists := target[key]
	if !exists {
		target[key] = value
		return
	}

	if arr, ok := current.(Array); ok {
		target[key] = append(arr, value)
		return
	}

	target[key] = Array{current, value}
}

func formatMultipart(result interface{}) (string, error) {
	obj, ok := result.(Object)
	if !ok {
		return "", unifiederrors.ValidationErrorf("multipart output expects an object, got %T", result)
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	if err := writer.SetBoundary(multipartBoundary); err != nil {
		return "", unifiederrors.WrapValidationf(err, "multipart boundary error: %v", err)
	}

	keys := make([]string, 0, len(obj))
	for key := range obj {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if err := writeMultipartValue(writer, key, obj[key]); err != nil {
			return "", err
		}
	}

	if err := writer.Close(); err != nil {
		return "", unifiederrors.WrapValidationf(err, "multipart close error: %v", err)
	}

	return buf.String(), nil
}

func writeMultipartValue(writer *multipart.Writer, key string, value interface{}) error {
	if arr, ok := value.(Array); ok {
		for _, item := range arr {
			if err := writeMultipartSingleValue(writer, key, item); err != nil {
				return err
			}
		}
		return nil
	}
	return writeMultipartSingleValue(writer, key, value)
}

func writeMultipartSingleValue(writer *multipart.Writer, key string, value interface{}) error {
	if obj, ok := value.(Object); ok && isMultipartStructuredPart(obj) {
		return writeMultipartStructuredPart(writer, key, obj)
	}
	return writer.WriteField(key, formatMultipartScalarValue(value))
}

func isMultipartStructuredPart(obj Object) bool {
	if _, ok := obj["content"]; ok {
		return true
	}
	if _, ok := obj["filename"]; ok {
		return true
	}
	if _, ok := obj["contentType"]; ok {
		return true
	}
	return false
}

func writeMultipartStructuredPart(writer *multipart.Writer, key string, obj Object) error {
	content := ""
	if rawContent, ok := obj["content"]; ok {
		content = formatMultipartScalarValue(rawContent)
	}

	headers := make(textproto.MIMEHeader)
	headers.Set("Content-Disposition", buildMultipartContentDisposition(key, obj))

	if rawType, ok := obj["contentType"]; ok {
		contentType := strings.TrimSpace(formatMultipartScalarValue(rawType))
		if contentType != "" {
			headers.Set("Content-Type", contentType)
		}
	}

	partWriter, err := writer.CreatePart(headers)
	if err != nil {
		return unifiederrors.WrapValidationf(err, "multipart part error: %v", err)
	}
	if _, err := io.WriteString(partWriter, content); err != nil {
		return unifiederrors.WrapValidationf(err, "multipart part write error: %v", err)
	}
	return nil
}

func buildMultipartContentDisposition(name string, obj Object) string {
	disposition := fmt.Sprintf(`form-data; name="%s"`, escapeMultipartQuoted(name))
	if rawFilename, ok := obj["filename"]; ok {
		filename := formatMultipartScalarValue(rawFilename)
		disposition += fmt.Sprintf(`; filename="%s"`, escapeMultipartQuoted(filename))
	}
	return disposition
}

func escapeMultipartQuoted(value string) string {
	return strings.ReplaceAll(value, `"`, `\"`)
}

func formatMultipartScalarValue(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return typed
	case nil:
		return ""
	default:
		return marshalToJSON(value)
	}
}

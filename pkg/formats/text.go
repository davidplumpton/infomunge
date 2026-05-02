package formats

func init() {
	RegisterReader("text/plain", readText)
	RegisterWriter("text/plain", formatText)
	RegisterExtension(".txt", "text/plain")
}

func readText(content string) (interface{}, error) {
	return content, nil
}

func formatText(result interface{}) (string, error) {
	return valueToPlainString(result), nil
}

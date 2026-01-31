package formats

func init() {
	RegisterReader("text/plain", readText)
	RegisterExtension(".txt", "text/plain")
}

func readText(content string) (interface{}, error) {
	return content, nil
}

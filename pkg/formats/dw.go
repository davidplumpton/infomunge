package formats

func init() {
	RegisterReader("application/dw", readDW)
	RegisterWriter("application/dw", formatBinary)
	RegisterExtension(".dw", "application/dw")
	RegisterExtension(".dwl", "application/dw")
}

func readDW(content string) (interface{}, error) {
	return content, nil
}

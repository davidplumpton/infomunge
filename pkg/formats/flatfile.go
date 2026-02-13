package formats

func init() {
	RegisterReader("application/flatfile", readBinary)
	RegisterWriter("application/flatfile", formatBinary)
	RegisterExtension(".flatfile", "application/flatfile")
	RegisterExtension(".ffd", "application/flatfile")
}

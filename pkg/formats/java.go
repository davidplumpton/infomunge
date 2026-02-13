package formats

func init() {
	RegisterReader("application/java", readBinary)
	RegisterWriter("application/java", formatBinary)
	RegisterExtension(".java", "application/java")
	RegisterExtension(".ser", "application/java")
}

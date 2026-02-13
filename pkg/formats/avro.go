package formats

func init() {
	RegisterReader("application/avro", readBinary)
	RegisterWriter("application/avro", formatBinary)
	RegisterExtension(".avro", "application/avro")
}

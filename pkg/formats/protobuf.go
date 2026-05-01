package formats

func init() {
	RegisterReader("application/protobuf", readBinary)
	RegisterWriter("application/protobuf", formatBinary)
	RegisterReader("application/x-protobuf", readBinary)
	RegisterWriter("application/x-protobuf", formatBinary)
	RegisterExtension(".protobuf", "application/protobuf")
	RegisterExtension(".pb", "application/protobuf")
	RegisterExtension(".pbf", "application/protobuf")
	RegisterOptionsAlias("application/x-protobuf", "application/protobuf")
	RegisterReadOptionsHandler("application/protobuf", readProtobufWithOptions)
	RegisterWriteOptionsHandler("application/protobuf", formatProtobufWithOptions)
}

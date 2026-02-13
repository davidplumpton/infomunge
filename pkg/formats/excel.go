package formats

func init() {
	RegisterReader("application/xlsx", readBinary)
	RegisterWriter("application/xlsx", formatBinary)
	RegisterReader("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", readBinary)
	RegisterWriter("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", formatBinary)
	RegisterExtension(".xlsx", "application/xlsx")
	RegisterExtension(".excel", "application/xlsx")
}

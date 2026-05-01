package formats

func init() {
	RegisterReader("application/xlsx", readBinary)
	RegisterWriter("application/xlsx", formatBinary)
	RegisterReader("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", readBinary)
	RegisterWriter("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", formatBinary)
	RegisterExtension(".xlsx", "application/xlsx")
	RegisterExtension(".excel", "application/xlsx")
	RegisterOptionsAlias("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "application/xlsx")
	RegisterReadOptionsHandler("application/xlsx", readXLSXWithOptions)
	RegisterWriteOptionsHandler("application/xlsx", formatXLSXWithOptions)
}

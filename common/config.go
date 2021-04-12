package common

import (
	"net/url"
)

const SYSTEM_FOLDER_PATH = "system"
const TEMP_DOWNLOAD_FOLDER_PATH = "temp"
const DEFAULT_UNDERTOW_ACCESS = "1816b8UsorvWzv7Dj8bs45pFAPaNCyniKFRsiPJuysKy8cWKkPt8hTAvVbbpyZeE3FFKRaWyru8S6U2ooR3YdfhybCYAZbNHWD4x88ZAWgFe6ScSaGBrnJ3Fde8zqXbf2EuZFBMJKunq99s9FkU9CuQAjoa5JqLR9BKdrjSdycG52Z1ZNpyDiYZ4m6rgnFEDoqXrgtnzV7SayywECsbokWFLcH4Ef92GZLZsVJ61QK8qbr1zXaKRNJ7E9R3R9c3DMA4ujZf2s5b5tknuqLhcx2bQfM1suoK7U7Tz32Dbr77BA38C9joMGMoF7JUoerHe7bv58tiRQ3Jgo8MgUd7rZmgEyisKtcjfpxfRvgUsqKhJunkPGQbaGxD3jvw5Htqjwx8MG9PbTPmwUPDv4ZgYZBWvd8Pva3UbiicJVWcjbLr4p98BEJRbMANB5WpX8Br6fB6jxs8QXxyWRnySATmWymbYHQtLt99kQmDVPkfhsaz"

func GetDefaultUndertowURL() url.URL {
	return url.URL{
		Scheme: "sj",
		Host:   "arkhive",
		Path:   "undertow.tow",
	}
}

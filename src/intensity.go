package src

var (
	basePathsSmall = []string{
		"backup",
		"backup/backup",
		"backups/backup",
	}
	basePathsMedium = []string{
		"backup",
		"backup/backup",
		"backups/backup",
		"www",
		"database",
		"log",
		"logs",
		"123",
	}
	basePathsBig = []string{
		"backup",
		"backup/backup",
		"backups/backup",
		"www",
		"database",
		"log",
		"logs",
		"db",
		"logs/log",
		"logs/logs",
		"site",
		"http",
		"website",
		"1",
		"123",
	}

	extensionsSmall = []string{
		"zip",
		"tar.gz",
	}
	extensionsMedium = []string{
		"zip",
		"tar.gz",
		"tar",
		"rar",
	}
	extensionsBig = []string{
		"zip",
		"tar.gz",
		"tar",
		"rar",
		"7z",
		"gz",
		"bz2",
	}
)

func GetBasePathsAndExtensions(intensity string) ([]string, []string) {
	switch intensity {
	case "small":
		return basePathsSmall, extensionsSmall
	case "big":
		return basePathsBig, extensionsBig
	case "medium":
		fallthrough // default
	default:
		return basePathsMedium, extensionsMedium
	}
}

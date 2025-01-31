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

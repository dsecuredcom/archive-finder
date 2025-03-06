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
		"api",
		"admin",
		"inc",
		"private",
		"private/backup",
	}
	basePathsBig = []string{
		"backup",
		"backup/backup",
		"backups/backup",
		"admin/backup",
		"admin/log",
		"admin/backups/backup",
		"admin/logs/log",
		"www",
		"html",
		"app",
		"bin",
		"config",
		"src",
		"files",
		"uploads",
		"database",
		"mysql",
		"sql",
		"log",
		"logs",
		"export",
		"snapshot",
		"old",
		"db",
		"logs/log",
		"logs/logs",
		"site",
		"http",
		"website",
		"1",
		"123",
		"api",
		"admin",
		"inc",
		"private",
		"private/backup",
		"dump",
		"users",
		"backup1",
		"backup2",
		"backup3",
		"admins",
		"dist", "assets", "js",
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

func GetBasePathsAndExtensions(config *Config) ([]string, []string) {
	var basePaths, extensions []string

	switch config.Intensity {
	case "small":
		basePaths, extensions = basePathsSmall, extensionsSmall
	case "big":
		basePaths, extensions = basePathsBig, extensionsBig
	case "medium":
		fallthrough
	default:
		basePaths, extensions = basePathsMedium, extensionsMedium
	}

	if len(config.UserBaseWords) > 0 {
		basePaths = config.UserBaseWords
	}
	if len(config.UserExtensions) > 0 {
		extensions = config.UserExtensions
	}

	return basePaths, extensions
}

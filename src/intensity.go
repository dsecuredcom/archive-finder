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
		"backup", "backup/backup", "backups/backup", "admin/backup", "backup1", "backup2", "backup3", "backup_full", "admin/backups/backup", "private/backup",
		"admin/log", "admin/logs/log", "log", "logs", "logs/log", "logs/logs",
		"www", "html", "site", "http", "website", "app", "apps", "main", "web", "server", "public",
		"bin", "bins", "dlls",
		"config", "conf", "configs", "cfg", "env", "configuration", "credentials", "creds",
		"src", "old", "inc", "data", "userdata",
		"files", "uploads", "upload",
		"database", "mysql", "sql", "snapshot", "db", "dump", "export", "exports",
		"report", "reports", "reporting",
		"1", "12", "123", "1234",
		"api", "admins", "admin", "private", "users", "user", "tests",
		"dist", "assets", "js", "script", "scripts", "resources",
		"prod", "dev", "test", "staging",
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

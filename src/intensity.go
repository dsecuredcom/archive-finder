package src

var (
	basePathsSmall = []string{
		"backup",
		"log",
		"logs",
	}

	backupFoldersSmall = []string{
		"backups",
		"logs",
	}

	basePathsMedium = []string{
		"backup",
		"www",
		"app",
		"database",
		"db",
		"log",
		"logs",
	}

	backupFoldersMedium = []string{
		"backups",
		"backup",
		"logs",
	}

	basePathsBig = []string{
		"backup", "backup1", "backup2", "backup3", "backup_full",
		"log", "logs",
		"www", "html", "site", "http", "httpd", "website", "app", "apps", "main", "web", "server", "public",
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
	backupFoldersBig = []string{
		"backups",
		"backup",
		"logs",
		"_backups",
		"_backup_",
		"old",
		"private",
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

func GetBasePathsAndExtensions(config *Config) ([]string, []string, []string) {
	var basePaths, extensions, folders []string

	switch config.Intensity {
	case "small":
		basePaths, extensions, folders = basePathsSmall, extensionsSmall, backupFoldersSmall
	case "big":
		basePaths, extensions, folders = basePathsBig, extensionsBig, backupFoldersBig
	case "medium":
		fallthrough
	default:
		basePaths, extensions, folders = basePathsMedium, extensionsMedium, backupFoldersMedium
	}

	if len(config.UserBaseWords) > 0 {
		basePaths = config.UserBaseWords
	}
	if len(config.UserExtensions) > 0 {
		extensions = config.UserExtensions
	}

	if len(config.BackupFolders) > 0 {
		folders = config.BackupFolders
	}

	return basePaths, extensions, folders
}

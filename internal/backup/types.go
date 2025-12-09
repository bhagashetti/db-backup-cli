package backup

// BackupOptions holds everything needed to perform a backup.
type BackupOptions struct {
	DBType   string
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	Output   string
}

// RestoreOptions holds everything needed to perform a restore.
type RestoreOptions struct {
	DBType   string
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	Input    string
}

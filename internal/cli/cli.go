package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bhagashetti/db-backup-cli/internal/backup"
	"github.com/bhagashetti/db-backup-cli/internal/config"
	"github.com/bhagashetti/db-backup-cli/internal/logs"
	"github.com/bhagashetti/db-backup-cli/internal/storage"
)

const appVersion = "0.2.0"

func Execute() {
	// init logging
	logFile, err := logs.Init("backup.log")
	if err != nil {
		fmt.Println("Failed to initialize logger:", err)
		os.Exit(1)
	}
	defer logFile.Close()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	logs.Info("Command received: %s", command)

	switch command {
	case "backup":
		handleBackup(os.Args[2:])
	case "restore":
		handleRestore(os.Args[2:])
	case "schedule":
		handleSchedule(os.Args[2:])
	case "version":
		fmt.Println("db-backup-cli version", appVersion)
		logs.Info("Version requested: %s", appVersion)
	case "help":
		printUsage()
		logs.Info("Help requested")
	default:
		fmt.Println("Unknown command:", command)
		printUsage()
		logs.Error("Unknown command: %s", command)
		os.Exit(1)
	}

}

func printUsage() {
	fmt.Println("Usage: db-backup-cli <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  backup     Run a backup")
	fmt.Println("  restore    Restore from a backup")
	fmt.Println("  schedule   Run backups on a fixed interval")
	fmt.Println("  version    Show application version")
	fmt.Println("  help       Show this help message")
	fmt.Println()
	fmt.Println("Use 'db-backup-cli <command> -h' to see options for a command.")
}

func handleBackup(args []string) {
	fs := flag.NewFlagSet("backup", flag.ExitOnError)

	configPath := fs.String("config", "", "Path to JSON config file")

	dbType := fs.String("db-type", "mysql", "Database type (mysql, postgres, mongo, sqlite)")
	host := fs.String("host", "localhost", "Database host")
	port := fs.Int("port", 3306, "Database port")
	user := fs.String("user", "root", "Database user")
	password := fs.String("password", "", "Database password")
	dbName := fs.String("db", "", "Database name")
	output := fs.String("out", "backup.sql", "Output backup file")
	compressFlag := fs.Bool("compress", false, "Compress backup using gzip (.gz)")
	encryptFlag := fs.Bool("encrypt", false, "Encrypt backup using AES-256-GCM")
	encryptKeyFlag := fs.String("encrypt-key", "", "Encryption key (32 chars)")

	fs.Parse(args)

	var (
		opts       backup.BackupOptions
		compress   bool
		encrypt    bool
		encryptKey string
		uploadS3   bool
		s3Bucket   string
		s3Region   string
		s3Prefix   string
	)

	// If a config file is provided, load values from it.
	if *configPath != "" {
		logs.Info("Loading backup config from file: %s", *configPath)
		cfg, err := config.LoadBackup(*configPath)
		if err != nil {
			fmt.Println("Failed to load config:", err)
			logs.Error("Failed to load backup config: %v", err)
			os.Exit(1)
		}

		opts = backup.BackupOptions{
			DBType:   cfg.DBType,
			Host:     cfg.Host,
			Port:     cfg.Port,
			User:     cfg.User,
			Password: cfg.Password,
			DBName:   cfg.DBName,
			Output:   cfg.Output,
		}
		compress = cfg.Compress
		encrypt = cfg.Encrypt
		encryptKey = cfg.EncryptKey
		uploadS3 = cfg.UploadS3
		s3Bucket = cfg.S3Bucket
		s3Region = cfg.S3Region
		s3Prefix = cfg.S3Prefix

		// If useTimestamp is true, change Output to include date-time.
		if cfg.UseTimestamp {
			timestamp := time.Now().Format("20060102-150405")
			opts.Output = fmt.Sprintf("%s-%s.sql", cfg.DBName, timestamp)
		}
	} else {
		// No config file: use CLI flags.
		if *dbName == "" {
			fmt.Println("Error: -db is required")
			fs.Usage()
			logs.Error("Backup failed: missing -db flag")
			os.Exit(1)
		}

		opts = backup.BackupOptions{
			DBType:   *dbType,
			Host:     *host,
			Port:     *port,
			User:     *user,
			Password: *password,
			DBName:   *dbName,
			Output:   *output,
		}
		compress = *compressFlag
		encrypt = *encryptFlag
		encryptKey = *encryptKeyFlag
		uploadS3 = false
		s3Bucket = ""
		s3Region = ""
		s3Prefix = ""

	}

	fmt.Println("Starting backup...")
	fmt.Printf("  db-type : %s\n", opts.DBType)
	fmt.Printf("  host    : %s\n", opts.Host)
	fmt.Printf("  port    : %d\n", opts.Port)
	fmt.Printf("  user    : %s\n", opts.User)
	fmt.Printf("  db      : %s\n", opts.DBName)
	fmt.Printf("  out     : %s\n", opts.Output)
	fmt.Printf("  compress: %v\n", compress)
	fmt.Printf("  encrypt : %v\n", encrypt)

	logs.Info(
		"Starting backup: dbType=%s host=%s port=%d user=%s db=%s out=%s compress=%v encrypt=%v",
		opts.DBType, opts.Host, opts.Port, opts.User, opts.DBName, opts.Output, compress, encrypt,
	)

	// 1) Run DB-specific backup
	switch opts.DBType {
	case "mysql":
		if err := backup.MySQLBackup(opts); err != nil {
			fmt.Println("Backup failed:", err)
			logs.Error("Backup failed: %v", err)
			os.Exit(1)
		}
	default:
		fmt.Println("Unsupported db-type for now:", opts.DBType)
		logs.Error("Unsupported db-type: %s", opts.DBType)
		os.Exit(1)
	}

	// Track the current "final" file path as we transform it
	finalPath := opts.Output

	// 2) Optional compression
	if compress {
		gzPath := finalPath + ".gz"
		fmt.Println("Compressing backup to:", gzPath)
		logs.Info("Compressing backup to: %s", gzPath)

		if err := backup.GzipFile(finalPath, gzPath); err != nil {
			fmt.Println("Compression failed:", err)
			logs.Error("Compression failed: %v", err)
			os.Exit(1)
		}

		if err := os.Remove(finalPath); err != nil {
			fmt.Println("Warning: could not remove original file:", err)
			logs.Error("Could not remove original backup file: %v", err)
		} else {
			logs.Info("Removed original uncompressed backup: %s", finalPath)
		}

		finalPath = gzPath
	}

	// 3) Optional encryption
	if encrypt {
		if encryptKey == "" {
			fmt.Println("Encryption requested but no key provided")
			logs.Error("Encryption requested but no key provided")
			os.Exit(1)
		}

		keyBytes := []byte(encryptKey)
		if len(keyBytes) != 32 {
			fmt.Println("Encryption key must be exactly 32 characters")
			logs.Error("Encryption key invalid length: %d", len(keyBytes))
			os.Exit(1)
		}

		encPath := finalPath + ".enc"
		fmt.Println("Encrypting backup to:", encPath)
		logs.Info("Encrypting backup to: %s", encPath)

		if err := backup.EncryptFile(finalPath, encPath, keyBytes); err != nil {
			fmt.Println("Encryption failed:", err)
			logs.Error("Encryption failed: %v", err)
			os.Exit(1)
		}

		if err := os.Remove(finalPath); err != nil {
			fmt.Println("Warning: could not remove unencrypted file:", err)
			logs.Error("Could not remove unencrypted file: %v", err)
		} else {
			logs.Info("Removed unencrypted backup: %s", finalPath)
		}

		finalPath = encPath
	}
	// 4) Optional S3 upload
	if uploadS3 {
		if s3Bucket == "" || s3Region == "" {
			fmt.Println("S3 upload requested but bucket or region is empty")
			logs.Error("S3 upload requested but bucket or region is empty")
			os.Exit(1)
		}

		key := s3Prefix + filepath.Base(finalPath)
		fmt.Println("Uploading backup to S3:", s3Bucket, "key:", key)
		logs.Info("Uploading backup to S3: bucket=%s key=%s region=%s", s3Bucket, key, s3Region)

		if err := storage.UploadToS3(s3Bucket, s3Region, key, finalPath); err != nil {
			fmt.Println("S3 upload failed:", err)
			logs.Error("S3 upload failed: %v", err)
			os.Exit(1)
		}

		fmt.Println("S3 upload completed.")
		logs.Info("S3 upload completed: bucket=%s key=%s", s3Bucket, key)
	}

	fmt.Println("Backup completed successfully. Final file:", finalPath)
	logs.Info("Backup completed successfully. Final file: %s", finalPath)
}

func handleRestore(args []string) {
	fs := flag.NewFlagSet("restore", flag.ExitOnError)

	configPath := fs.String("config", "", "Path to JSON restore config file")

	dbType := fs.String("db-type", "mysql", "Database type")
	host := fs.String("host", "localhost", "Database host")
	port := fs.Int("port", 3306, "Database port")
	user := fs.String("user", "root", "Database user")
	password := fs.String("password", "", "Database password")
	dbName := fs.String("db", "", "Database name")
	input := fs.String("in", "backup.sql", "Backup file to restore from")

	fs.Parse(args)

	var opts backup.RestoreOptions

	if *configPath != "" {
		logs.Info("Loading restore config from file: %s", *configPath)
		cfg, err := config.LoadRestore(*configPath)
		if err != nil {
			fmt.Println("Failed to load restore config:", err)
			logs.Error("Failed to load restore config: %v", err)
			os.Exit(1)
		}

		opts = backup.RestoreOptions{
			DBType:   cfg.DBType,
			Host:     cfg.Host,
			Port:     cfg.Port,
			User:     cfg.User,
			Password: cfg.Password,
			DBName:   cfg.DBName,
			Input:    cfg.Input,
		}
	} else {
		if *dbName == "" {
			fmt.Println("Error: -db is required")
			fs.Usage()
			logs.Error("Restore failed: missing -db flag")
			os.Exit(1)
		}

		opts = backup.RestoreOptions{
			DBType:   *dbType,
			Host:     *host,
			Port:     *port,
			User:     *user,
			Password: *password,
			DBName:   *dbName,
			Input:    *input,
		}
	}

	fmt.Println("Starting restore...")
	fmt.Printf("  db-type: %s\n", opts.DBType)
	fmt.Printf("  host   : %s\n", opts.Host)
	fmt.Printf("  port   : %d\n", opts.Port)
	fmt.Printf("  user   : %s\n", opts.User)
	fmt.Printf("  db     : %s\n", opts.DBName)
	fmt.Printf("  in     : %s\n", opts.Input)

	logs.Info(
		"Starting restore: dbType=%s host=%s port=%d user=%s db=%s in=%s",
		opts.DBType, opts.Host, opts.Port, opts.User, opts.DBName, opts.Input,
	)

	switch opts.DBType {
	case "mysql":
		if err := backup.MySQLRestore(opts); err != nil {
			fmt.Println("Restore failed:", err)
			logs.Error("Restore failed: %v", err)
			os.Exit(1)
		}
		fmt.Println("Restore completed successfully.")
		logs.Info("Restore completed successfully.")
	default:
		fmt.Println("Unsupported db-type for now:", opts.DBType)
		logs.Error("Unsupported db-type: %s", opts.DBType)
		os.Exit(1)
	}
}

func handleSchedule(args []string) {
	fs := flag.NewFlagSet("schedule", flag.ExitOnError)

	configPath := fs.String("config", "", "Path to JSON backup config file")
	every := fs.String("every", "", "How often to run the backup (e.g. 1h, 30m, 24h)")
	daily := fs.String("daily", "", "Run backup once per day at HH:MM (24h format, local time)")

	fs.Parse(args)

	if *configPath == "" {
		fmt.Println("Error: -config is required for schedule")
		fs.Usage()
		logs.Error("Schedule failed: missing -config flag")
		os.Exit(1)
	}

	if *every == "" && *daily == "" {
		fmt.Println("Error: either -every or -daily must be provided")
		fs.Usage()
		logs.Error("Schedule failed: missing -every/-daily")
		os.Exit(1)
	}

	if *every != "" && *daily != "" {
		fmt.Println("Error: use either -every OR -daily, not both")
		fs.Usage()
		logs.Error("Schedule failed: both -every and -daily provided")
		os.Exit(1)
	}

	// Interval scheduler: -every=1h
	if *every != "" {
		interval, err := time.ParseDuration(*every)
		if err != nil {
			fmt.Println("Invalid duration for -every:", err)
			logs.Error("Invalid duration for -every: %v", err)
			os.Exit(1)
		}

		fmt.Println("Starting interval scheduler...")
		fmt.Println("  config :", *configPath)
		fmt.Println("  every  :", interval)

		logs.Info("Starting interval scheduler: config=%s every=%s", *configPath, interval.String())

		for {
			fmt.Println("Running scheduled backup at", time.Now().Format(time.RFC3339))
			logs.Info("Running scheduled backup at %s", time.Now().Format(time.RFC3339))

			handleBackup([]string{"-config=" + *configPath})

			fmt.Println("Next backup in:", interval)
			logs.Info("Next backup in: %s", interval.String())

			time.Sleep(interval)
		}
	}

	// Daily scheduler: -daily=HH:MM
	t, err := time.Parse("15:04", *daily)
	if err != nil {
		fmt.Println("Invalid time for -daily (expected HH:MM):", err)
		logs.Error("Invalid time for -daily: %v", err)
		os.Exit(1)
	}

	fmt.Println("Starting daily scheduler...")
	fmt.Println("  config :", *configPath)
	fmt.Println("  daily  :", *daily)

	logs.Info("Starting daily scheduler: config=%s daily=%s", *configPath, *daily)

	for {
		now := time.Now()
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())
		if !nextRun.After(now) {
			// if time already passed today, schedule for tomorrow
			nextRun = nextRun.Add(24 * time.Hour)
		}

		wait := time.Until(nextRun)
		fmt.Println("Next backup at:", nextRun.Format(time.RFC3339))
		logs.Info("Next daily backup at: %s (in %s)", nextRun.Format(time.RFC3339), wait.String())

		time.Sleep(wait)

		fmt.Println("Running daily scheduled backup at", time.Now().Format(time.RFC3339))
		logs.Info("Running daily scheduled backup at %s", time.Now().Format(time.RFC3339))

		handleBackup([]string{"-config=" + *configPath})
	}
}

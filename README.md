📦 DB Backup CLI – Cross-Platform Database Backup Utility

A powerful, secure, and flexible command-line utility for backing up and restoring databases.
Supports MySQL (with easy extension for PostgreSQL, MongoDB), automated scheduling, compression, encryption, cloud upload to AWS S3, and detailed logging.

Built with Go using a clean modular architecture.

🚀 Features
✅ Database Backup

Full backup of MySQL databases using mysqldump

Custom DB connection settings (host, port, user, password, db name)

File naming with timestamps for versioning

✅ Backup Enhancements
Feature	Description
Compression	.sql → .sql.gz using gzip
Encryption	AES-256-GCM secure encrypted backup (.enc)
Log Rotation	Automatically rotates logs every 5MB
Config File Support	Run backup using config.json
Error Handling	Clear error messages with logging
✅ Restore Support

Restore .sql dumps back into MySQL

Validates DB connectivity

Clean restore command:

db-backup-cli restore -config=restore-config.json

✅ Scheduling

Automate backups:

Interval Scheduling
db-backup-cli schedule -config=config.json -every=1h

Daily Time Scheduling
db-backup-cli schedule -config=config.json -daily=02:00


Keeps running indefinitely like a lightweight cron job.

✅ AWS S3 Cloud Upload

After backup:

Upload the final encrypted file to an S3 bucket

Uses AWS SDK v2

Bucket prefix support (mysql-backups/your-file.enc)

Requires AWS credentials set in:

C:\Users\<user>\.aws\credentials
C:\Users\<user>\.aws\config


Example:

credentials

[default]
aws_access_key_id = YOUR_KEY_ID
aws_secret_access_key = YOUR_SECRET


config

[default]
region = ap-south-1

📁 Project Structure
db-backup-cli/
│
├── cmd/
│   └── db-backup-cli/
│       └── main.go       # CLI entry point
│
├── internal/
│   ├── backup/           # Backup logic (mysqldump, compression, encryption)
│   ├── config/           # Load config.json
│   ├── logs/             # Logging + rotation
│   ├── storage/          # AWS S3 upload
│   └── cli/              # CLI command handlers
│
├── config.json           # User backup config
├── restore-config.json   # Restore config example
└── README.md

🛠 Installation
1. Install Go

https://go.dev/dl/

2. Clone repository
git clone https://github.com/<yourusername>/db-backup-cli
cd db-backup-cli

3. Build executable
go build -o db-backup-cli.exe ./cmd/db-backup-cli

⚙️ Configuration

Example config.json:

{
  "dbType": "mysql",
  "host": "localhost",
  "port": 3306,
  "user": "root",
  "password": "YOURPASSWORD",
  "dbName": "backup_demo",
  "out": "backup_demo.sql",
  "compress": true,
  "useTimestamp": true,
  "encrypt": true,
  "encryptKey": "0123456789abcdef0123456789abcdef",
  "uploadS3": true,
  "s3Bucket": "db-backups-bhagash",
  "s3Region": "ap-south-1",
  "s3Prefix": "mysql-backups/"
}

🧪 Usage Examples
▶ Backup (using config)
db-backup-cli backup -config=config.json

▶ Restore (using config)
db-backup-cli restore -config=restore-config.json

▶ Schedule daily backup
db-backup-cli schedule -config=config.json -daily=02:00

▶ Schedule every 1 hour
db-backup-cli schedule -config=config.json -every=1h

🔐 Encryption Details

The tool uses:

AES-256 (GCM mode)

32-byte encryption key from config

Nonce automatically generated

Output extension: .enc

Decryption happens automatically during restore (if needed in future upgrades).

☁ AWS S3 Upload Details

After encryption, the file:

backup_demo-2025_timestamp.sql.gz.enc


is uploaded to:

s3://db-backups-bhagash/mysql-backups/

Requirements:

AWS IAM User with AmazonS3FullAccess

Credentials file in:

C:\Users\<username>\.aws\credentials


Region file in:

C:\Users\<username>\.aws\config

🧱 Future Enhancements (Optional)

Support PostgreSQL pg_dump

Support MongoDB mongodump

Restore directly from S3

Web dashboard for viewing backup history

Add decryption + restore for encrypted files

Slack/Email notifications after backup

🙌 Author

Anita Bhagashetti
Student Project • Database + Cloud + Go Programming

⭐ If this project helped you…

Leave a ⭐ star on GitHub!
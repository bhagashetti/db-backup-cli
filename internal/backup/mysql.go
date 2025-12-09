package backup

import (
	"fmt"
	"os"
	"os/exec"
)

// MySQLBackup performs a backup using mysqldump.
func MySQLBackup(opts BackupOptions) error {
	args := []string{
		"-h", opts.Host,
		"-P", fmt.Sprint(opts.Port),
		"-u", opts.User,
	}

	if opts.Password != "" {
		args = append(args, "-p"+opts.Password)
	}

	args = append(args, opts.DBName)

	cmd := exec.Command("mysqldump", args...)

	fmt.Println("Running command:", "mysqldump", args)

	outfile, err := os.Create(opts.Output)
	if err != nil {
		return fmt.Errorf("could not create output file: %w", err)
	}
	defer outfile.Close()

	cmd.Stdout = outfile
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mysqldump failed: %w", err)
	}

	return nil
}

// MySQLRestore restores a backup using mysql.
func MySQLRestore(opts RestoreOptions) error {
	args := []string{
		"-h", opts.Host,
		"-P", fmt.Sprint(opts.Port),
		"-u", opts.User,
	}

	if opts.Password != "" {
		args = append(args, "-p"+opts.Password)
	}

	args = append(args, opts.DBName)

	cmd := exec.Command("mysql", args...)

	fmt.Println("Running command:", "mysql", args)

	infile, err := os.Open(opts.Input)
	if err != nil {
		return fmt.Errorf("could not open input file: %w", err)
	}
	defer infile.Close()

	cmd.Stdin = infile
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mysql restore failed: %w", err)
	}

	return nil
}

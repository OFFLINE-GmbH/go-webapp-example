package daemon

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"go-webapp-example/pkg/fs"
	"go-webapp-example/pkg/log"

	"github.com/pkg/errors"
)

// Backup daemon dumps the database at a given time.
type Backup struct {
	Host      string
	Port      string
	Username  string
	Password  string
	Database  string
	TargetDir string
	Time      string
}

// NewBackup returns a new backup daemon.
func NewBackup(host, port, username, password, database, dir, runAt string) *Backup {
	return &Backup{
		Host:      host,
		Port:      port,
		Username:  username,
		Password:  password,
		Database:  database,
		TargetDir: dir,
		Time:      runAt,
	}
}

func (i *Backup) Name() string { return "backup" }

// Backup daemon handles incoming tcp connections.
func (i *Backup) Run(ctx context.Context, wg *sync.WaitGroup, logger log.Logger) error {
	wg.Add(1)
	defer wg.Done()

	go func() {
		for {
			select {
			case <-time.After(time.Minute):
				now := time.Now()
				if now.Format("15:04") != i.Time {
					fields := log.Fields{"now": now.Format("15:04"), "scheduled": i.Time}
					logger.WithFields(fields).Tracef("skipping mysql dump")
					continue
				}
				logger.Info("creating database dump")
				err := i.createBackupDump(ctx, logger)
				if err != nil {
					logger.Errorf("backup dump failed: %s", err)
					continue
				}
				logger.Info("database dump created successfully")
			case <-ctx.Done():
				logger.Debug("backup daemon is shutting down...")
				return
			}
		}
	}()

	// Wait for the daemon to stop
	<-ctx.Done()

	return nil
}

// nolint:funlen
func (i *Backup) createBackupDump(ctx context.Context, logger log.Logger) error {
	targetFile := filepath.Join(i.TargetDir, "..", "backups", time.Now().Format("2006-01-02"), "database.sql.gz")
	target, err := filepath.Abs(targetFile)
	if err != nil {
		return errors.WithStack(err)
	}
	err = fs.EnsureDir(filepath.Dir(target))
	if err != nil {
		return errors.WithStack(err)
	}
	// open the out file for writing
	outfile, err := os.Create(target)
	if err != nil {
		return errors.WithStack(err)
	}
	defer outfile.Close()
	// nolint:gosec
	mysqldump := exec.CommandContext(
		ctx,
		"mysqldump",
		"--single-transaction=true", // Don't lock the database during dump.
		fmt.Sprintf("-u%s", i.Username),
		fmt.Sprintf("-p%s", i.Password),
		fmt.Sprintf("-h%s", i.Host),
		fmt.Sprintf("-P%s", i.Port),
		i.Database,
	)
	gzip := exec.CommandContext(
		ctx,
		"gzip",
		"-9",
	)

	pr, pw := io.Pipe()
	mysqldump.Stdout = pw
	gzip.Stdin = pr
	gzip.Stdout = outfile

	logger.Debug("started mysqldump process")
	if err = mysqldump.Start(); err != nil {
		return errors.Wrap(err, "failed to run mysqldump process")
	}
	if err := gzip.Start(); err != nil {
		return errors.Wrap(err, "failed to start gzip process")
	}
	go func() {
		defer pw.Close()
		_ = mysqldump.Wait()
	}()
	if err := gzip.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				logger.Errorf("dump process exited with status: %d", status.ExitStatus())
			} else {
				logger.Errorf("dump process exited with unknown status: %s", err)
			}
			return exitErr
		}
		logger.Debugf("dump process exited with status: %s", err)
		return err
	}
	logger.Debug("dump process exited successfully")
	return nil
}

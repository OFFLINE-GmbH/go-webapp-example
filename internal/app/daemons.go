package app

import (
	"go-webapp-example/internal/daemon"
)

// StartDaemons starts all the application's background jobs.
func (k *Kernel) StartDaemons() {
	m := daemon.NewManager(k.context.ctx, k.wg, k.Log)

	go m.Start(daemon.NewExampleDaemon())

	if k.Config.Database.Backup {
		go m.Start(daemon.NewBackup(
			k.Config.Database.Host,
			k.Config.Database.Port,
			k.Config.Database.Username,
			k.Config.Database.Password,
			k.Config.Database.Name,
			k.Config.Server.StorageDir,
			k.Config.Database.BackupTime,
		))
	} else {
		k.Log.WithPrefix("dmn.backup").Info("database backups are disabled")
	}
}

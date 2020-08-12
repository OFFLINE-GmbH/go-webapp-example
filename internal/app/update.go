package app

import (
	"context"
	"os"
	"path/filepath"

	"go-webapp-example/pkg/log"

	"github.com/pkg/errors"
	"github.com/romanyx/polluter"
)

const VersionParam = "update_version"

type Updater struct {
	k   *Kernel
	log log.Logger
	// updates contains all update routines.
	updates []updateFn
}

// updateFn contains update logic.
type updateFn func(log log.Logger, k *Kernel) error

// NewUpdater returns a new updater instance.
func NewUpdater(k *Kernel) *Updater {
	u := &Updater{k: k, log: k.Log.WithPrefix("app.updater")}
	u.register()
	return u
}

func (u *Updater) register() {
	u.updates = []updateFn{
		seedBaseData, // 0
		someUpdate, // 1
		someOhterUpdate, // 2
	}
}

// seedBaseData seeds the base data for a new installation.
func seedBaseData(l log.Logger, k *Kernel) error {
	l.Println("seeding base data")
	// Seed
	seed, err := os.Open(filepath.Join(k.Config.Database.Migrations, "..", "seeds", "base.yml"))
	if err != nil {
		return errors.WithStack(err)
	}
	defer seed.Close()
	p := polluter.New(polluter.MySQLEngine(k.DB.Connection()))
	if err = p.Pollute(seed); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// someUpdate is here to show that this method is only run once when starting the server.
func someUpdate(l log.Logger, k *Kernel) error {
	l.Println("installing some update...")
	return nil
}

// someOhterUpdate is here to show that this method is only run once when starting the server.
func someOhterUpdate(l log.Logger, k *Kernel) error {
	l.Println("installing some other update...")
	return nil
}

// Run applies all pending updates. The VersionParam value from the database indicates
// what updates are already applied and what updates are pending.
func (u *Updater) Run(ctx context.Context) error {
	// resultingError is the error that is returned from this function if any update function fails.
	var resultingError error
	current, err := u.getCurrentVersion(ctx)
	if err != nil {
		return errors.WithStack(err)
	}
	initial := current
	u.log.Debugf("current %s is %d", VersionParam, initial)
	for version, update := range u.updates {
		// Skip already applied updates.
		if version <= current {
			continue
		}
		err = update(u.log, u.k)
		if err != nil {
			// Store the error and exit the update routine. The last successfully applied update
			// will be marked as the current "update_version". Failed ones are retried.
			resultingError = errors.Wrapf(err, "failed to apply update version %d", version)
			break
		}
		current = version
	}
	if initial == current && resultingError == nil {
		u.log.Info("nothing to update")
		return nil
	}
	u.log.Infof("updated to version %d", current)
	// Update the current version if everything went well.
	err = u.setCurrentVersion(ctx, current)
	if err != nil {
		return errors.WithStack(err)
	}
	return resultingError
}

func (u *Updater) getCurrentVersion(ctx context.Context) (int, error) {
	var current int
	err := u.k.DB.GetContext(ctx, &current, "SELECT value FROM `systemparams` WHERE param = ? LIMIT 1", VersionParam)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return -1, nil
		}
		return 0, errors.WithStack(err)
	}
	return current, nil
}

func (u *Updater) setCurrentVersion(ctx context.Context, current int) error {
	_, err := u.k.DB.ExecContext(ctx, "INSERT INTO systemparams (param, value) VALUES (?, ?) ON DUPLICATE KEY UPDATE value = ?", VersionParam, current, current)
	return errors.WithStack(err)
}

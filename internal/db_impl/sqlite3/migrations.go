package sqlite3

import (
	"context"
	"errors"
	"fmt"
	"github.com/ProtonMail/gluon/imap"
	v1 "github.com/ProtonMail/gluon/internal/db_impl/sqlite3/v1"

	"github.com/ProtonMail/gluon/db"
	"github.com/ProtonMail/gluon/internal/db_impl/sqlite3/utils"
	v0 "github.com/ProtonMail/gluon/internal/db_impl/sqlite3/v0"
	"github.com/sirupsen/logrus"
)

type Migration interface {
	Run(ctx context.Context, tx utils.QueryWrapper, generator imap.UIDValidityGenerator) error
}

var migrationList = []Migration{
	&v0.Migration{},
	&v1.Migration{},
}

func RunMigrations(ctx context.Context, tx utils.QueryWrapper, generator imap.UIDValidityGenerator) error {
	dbVersion, err := getDatabaseVersion(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to get db version: %w", err)
	}

	if dbVersion < 0 {
		logrus.Debug("Version table does not exist, running all migrations")

		for idx, m := range migrationList {
			logrus.Debugf("Running migration for version %v", idx)

			if err := m.Run(ctx, tx, generator); err != nil {
				return fmt.Errorf("failed to run migration %v: %w", idx, err)
			}
		}

		if err := updateDBVersion(ctx, tx, len(migrationList)-1); err != nil {
			return fmt.Errorf("failed to update db version:%w", err)
		}

		logrus.Debug("Migrations completed")

		return nil
	}

	logrus.Debugf("DB Version is %v", dbVersion)

	for i := dbVersion + 1; i < len(migrationList); i++ {
		logrus.Debugf("Running migration for version %v", i)

		if err := migrationList[i].Run(ctx, tx, generator); err != nil {
			return err
		}
	}

	if err := updateDBVersion(ctx, tx, len(migrationList)-1); err != nil {
		return fmt.Errorf("failed to update db version:%w", err)
	}

	logrus.Debug("Migrations completed")

	return nil
}

// getDatabaseVersion returns -1 if the version table does not exist or the  version information contained within.
func getDatabaseVersion(ctx context.Context, tx utils.QueryWrapper) (int, error) {
	query := "SELECT `name` FROM sqlite_master WHERE `type` = 'table' AND `name` NOT LIKE 'sqlite_%' AND `name` = 'gluon_version'"

	_, err := utils.MapQueryRow[string](ctx, tx, query)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return -1, nil
		}

		return 0, err
	}

	versionQuery := "SELECT `version` FROM gluon_version WHERE `id` = 0"

	return utils.MapQueryRow[int](ctx, tx, versionQuery)
}

func updateDBVersion(ctx context.Context, tx utils.QueryWrapper, version int) error {
	query := "UPDATE gluon_version SET `version` = ? WHERE `id` = 0"

	return utils.ExecQueryAndCheckUpdatedNotZero(ctx, tx, query, version)
}

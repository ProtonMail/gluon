package v2

import (
	"context"
	"fmt"
	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/db_impl/sqlite3/utils"
)

type Migration struct{}

func (m Migration) Run(ctx context.Context, tx utils.QueryWrapper, _ imap.UIDValidityGenerator) error {
	query := fmt.Sprintf("CREATE TABLE `%v` (`%v` INTEGER NOT NULL PRIMARY KEY , `%v` TEXT)",
		ConnectorSettingsTableName,
		ConnectorSettingsFieldID,
		ConnectorSettingsFieldValue,
	)

	if _, err := utils.ExecQuery(ctx, tx, query); err != nil {
		return fmt.Errorf("failed to create connector settings table: %w", err)
	}

	query = fmt.Sprintf(
		"INSERT INTO %v (`%v`, `%v`) VALUES (?,NULL)",
		ConnectorSettingsTableName,
		ConnectorSettingsFieldID,
		ConnectorSettingsFieldValue,
	)

	if _, err := utils.ExecQuery(ctx, tx, query, ConnectorSettingsDefaultID); err != nil {
		return fmt.Errorf("failed to create default connector settings entry: %w", err)
	}

	return nil
}

package sqlite3

import (
	"context"
	"fmt"

	"github.com/bradenaw/juniper/xmaps"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
)

type MigrationV0 struct{}

func (m MigrationV0) Run(ctx context.Context, tx TXWrapper) error {
	tables := []Table{
		&DeletedSubscriptionsTable{},
		&MailboxesTable{},
		&MailboxFlagsTable{},
		&MailboxAttrTable{},
		&MailboxPermFlagsTable{},
		&MessagesTable{},
		&MessageFlagsTable{},
		&UIDsTable{},
		&GluonVersionTable{},
	}

	tablesNames := xslices.Map(tables, func(t Table) string {
		return t.Name()
	})

	query := fmt.Sprintf("SELECT `name` FROM sqlite_master WHERE `type` = 'table' AND `name` NOT LIKE 'sqlite_%%' AND `name` IN (%v)",
		GenSQLIn(len(tables)))

	args := MapSliceToAny(tablesNames)

	sqlTables, err := MapQueryRows[string](ctx, tx, query, args...)
	if err != nil {
		return err
	}

	tablesSet := xmaps.SetFromSlice(sqlTables)

	for _, table := range tables {
		if !tablesSet.Contains(table.Name()) {
			logrus.Debugf("Table '%v' does not exist, creating", table.Name())

			if err := table.Create(ctx, tx); err != nil {
				return err
			}
		}
	}

	return nil
}

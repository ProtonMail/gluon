package v0

import (
	"context"
	"fmt"
	"github.com/ProtonMail/gluon/imap"

	"github.com/ProtonMail/gluon/internal/db_impl/sqlite3/utils"
	"github.com/bradenaw/juniper/xmaps"
	"github.com/bradenaw/juniper/xslices"
	"github.com/sirupsen/logrus"
)

type Migration struct{}

func (m Migration) Run(ctx context.Context, tx utils.QueryWrapper, _ imap.UIDValidityGenerator) error {
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
		utils.GenSQLIn(len(tables)))

	args := utils.MapSliceToAny(tablesNames)

	sqlTables, err := utils.MapQueryRows[string](ctx, tx, query, args...)
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

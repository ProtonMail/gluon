package v0

import (
	"context"

	"github.com/ProtonMail/gluon/internal/db_impl/sqlite3/utils"
)

type Table interface {
	Name() string
	Create(ctx context.Context, tx utils.QueryWrapper) error
}

func execQueries(ctx context.Context, tx utils.QueryWrapper, queries []string) error {
	for _, q := range queries {
		if _, err := utils.ExecQuery(ctx, tx, q); err != nil {
			return err
		}
	}

	return nil
}

type DeletedSubscriptionsTable struct{}

func (d DeletedSubscriptionsTable) Name() string {
	return "deleted_subscriptions"
}

func (d DeletedSubscriptionsTable) Create(ctx context.Context, tx utils.QueryWrapper) error {
	queries := []string{
		"CREATE TABLE `deleted_subscriptions` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `name` text NOT NULL, `remote_id` text NOT NULL)",
		"CREATE UNIQUE INDEX `deleted_subscriptions_name_key` ON `deleted_subscriptions` (`name`)",
		"CREATE UNIQUE INDEX `deleted_subscriptions_remote_id_key` ON `deleted_subscriptions` (`remote_id`)",
		"CREATE INDEX `deletedsubscription_remote_id` ON `deleted_subscriptions` (`remote_id`)",
		"CREATE INDEX `deletedsubscription_name` ON `deleted_subscriptions` (`name`)",
	}

	return execQueries(ctx, tx, queries)
}

type MailboxesTable struct{}

func (m MailboxesTable) Name() string {
	return "mailboxes"
}

func (m MailboxesTable) Create(ctx context.Context, tx utils.QueryWrapper) error {
	queries := []string{
		"CREATE TABLE `mailboxes` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `remote_id` text NULL, `name` text NOT NULL, `uid_next` integer NOT NULL DEFAULT 1, `uid_validity` integer NOT NULL DEFAULT 1, `subscribed` bool NOT NULL DEFAULT true)",
		"CREATE UNIQUE INDEX `mailboxes_remote_id_key` ON `mailboxes` (`remote_id`)",
		"CREATE UNIQUE INDEX `mailboxes_name_key` ON `mailboxes` (`name`)",
		"CREATE INDEX `mailbox_id` ON `mailboxes` (`id`)",
		"CREATE INDEX `mailbox_remote_id` ON `mailboxes` (`remote_id`)",
		"CREATE INDEX `mailbox_name` ON `mailboxes` (`name`)",
	}

	return execQueries(ctx, tx, queries)
}

type MailboxAttrTable struct{}

func (m MailboxAttrTable) Name() string {
	return "mailbox_attrs"
}

func (m MailboxAttrTable) Create(ctx context.Context, tx utils.QueryWrapper) error {
	queries := []string{
		"CREATE TABLE `mailbox_attrs` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `value` text NOT NULL, `mailbox_attributes` integer NULL, CONSTRAINT `mailbox_attrs_mailboxes_attributes` FOREIGN KEY (`mailbox_attributes`) REFERENCES `mailboxes` (`id`) ON DELETE CASCADE)",
	}

	return execQueries(ctx, tx, queries)
}

type MailboxFlagsTable struct{}

func (m MailboxFlagsTable) Name() string {
	return "mailbox_flags"
}

func (m MailboxFlagsTable) Create(ctx context.Context, tx utils.QueryWrapper) error {
	queries := []string{
		"CREATE TABLE `mailbox_flags` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `value` text NOT NULL, `mailbox_flags` integer NULL, CONSTRAINT `mailbox_flags_mailboxes_flags` FOREIGN KEY (`mailbox_flags`) REFERENCES `mailboxes` (`id`) ON DELETE CASCADE)",
	}

	return execQueries(ctx, tx, queries)
}

type MailboxPermFlagsTable struct{}

func (m MailboxPermFlagsTable) Name() string {
	return "mailbox_perm_flags"
}

func (m MailboxPermFlagsTable) Create(ctx context.Context, tx utils.QueryWrapper) error {
	queries := []string{
		"CREATE TABLE `mailbox_perm_flags` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `value` text NOT NULL, `mailbox_permanent_flags` integer NULL, CONSTRAINT `mailbox_perm_flags_mailboxes_permanent_flags` FOREIGN KEY (`mailbox_permanent_flags`) REFERENCES `mailboxes` (`id`) ON DELETE CASCADE)",
	}

	return execQueries(ctx, tx, queries)
}

type MessagesTable struct{}

func (m MessagesTable) Name() string {
	return "messages"
}

func (m MessagesTable) Create(ctx context.Context, tx utils.QueryWrapper) error {
	queries := []string{
		"CREATE TABLE `messages` (`id` uuid NOT NULL, `remote_id` text NULL, `date` datetime NOT NULL, `size` integer NOT NULL, `body` text NOT NULL, `body_structure` text NOT NULL, `envelope` text NOT NULL, `deleted` bool NOT NULL DEFAULT false, PRIMARY KEY (`id`))",
		"CREATE UNIQUE INDEX `messages_remote_id_key` ON `messages` (`remote_id`)",
		"CREATE INDEX `message_id` ON `messages` (`id`)",
		"CREATE INDEX `message_remote_id` ON `messages` (`remote_id`)",
	}

	return execQueries(ctx, tx, queries)
}

type MessageFlagsTable struct{}

func (m MessageFlagsTable) Name() string {
	return "message_flags"
}

func (m MessageFlagsTable) Create(ctx context.Context, tx utils.QueryWrapper) error {
	queries := []string{
		"CREATE TABLE `message_flags` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `value` text NOT NULL, `message_flags` uuid NULL, CONSTRAINT `message_flags_messages_flags` FOREIGN KEY (`message_flags`) REFERENCES `messages` (`id`) ON DELETE CASCADE)",
	}

	return execQueries(ctx, tx, queries)
}

type UIDsTable struct{}

func (U UIDsTable) Name() string {
	return "ui_ds"
}

func (U UIDsTable) Create(ctx context.Context, tx utils.QueryWrapper) error {
	queries := []string{
		"CREATE TABLE `ui_ds` (`id` integer NOT NULL PRIMARY KEY AUTOINCREMENT, `uid` integer NOT NULL, `deleted` bool NOT NULL DEFAULT false, `recent` bool NOT NULL DEFAULT true, `mailbox_ui_ds` integer NULL, `uid_message` uuid NULL, CONSTRAINT `ui_ds_mailboxes_UIDs` FOREIGN KEY (`mailbox_ui_ds`) REFERENCES `mailboxes` (`id`) ON DELETE CASCADE, CONSTRAINT `ui_ds_messages_message` FOREIGN KEY (`uid_message`) REFERENCES `messages` (`id`) ON DELETE SET NULL)",
		"CREATE INDEX `uid_uid_uid_message` ON `ui_ds` (`uid`, `uid_message`)",
	}

	return execQueries(ctx, tx, queries)
}

type GluonVersionTable struct{}

func (g GluonVersionTable) Name() string {
	return "gluon_version"
}

func (g GluonVersionTable) Create(ctx context.Context, tx utils.QueryWrapper) error {
	queries := []string{
		"CREATE TABLE `gluon_version` (`id` integer NOT NULL PRIMARY KEY CHECK(`id` =0), `version` integer NOT NULL)",
		"INSERT INTO gluon_version (`id`, `version`) VALUES (0,0)",
	}

	return execQueries(ctx, tx, queries)
}

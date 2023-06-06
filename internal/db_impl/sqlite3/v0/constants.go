package v0

const DeletedSubscriptionsTableName = "deleted_subscriptions"
const DeletedSubscriptionsFieldName = "name"
const DeletedSubscriptionsFieldRemoteID = "remote_id"

const MailboxAttrsTableName = "mailbox_attrs"
const MailboxAttrsFieldValue = "value"
const MailboxAttrsFieldMailboxID = "mailbox_attributes"

const MailboxFlagsTableName = "mailbox_flags"
const MailboxFlagsFieldValue = "value"
const MailboxFlagsFieldMailboxID = "mailbox_flags"

const MailboxPermFlagsTableName = "mailbox_perm_flags"
const MailboxPermFlagsFieldValue = "value"
const MailboxPermFlagsFieldMailboxID = "mailbox_permanent_flags"

const MailboxesTableName = "mailboxes"
const MailboxesFieldID = "id"
const MailboxesFieldRemoteID = "remote_id"
const MailboxesFieldName = "name"
const MailboxesFieldUIDNext = "uid_next"
const MailboxesFieldUIDValidity = "uid_validity"
const MailboxesFieldSubscribed = "subscribed"

const MessageFlagsTableName = "message_flags"
const MessageFlagsFieldValue = "value"
const MessageFlagsFieldMessageID = "message_flags"

const MessagesTableName = "messages"
const MessagesFieldID = "id"
const MessagesFieldRemoteID = "remote_id"
const MessagesFieldDate = "date"
const MessagesFieldSize = "size"
const MessagesFieldBody = "body"
const MessagesFieldBodyStructure = "body_structure"
const MessagesFieldEnvelope = "envelope"
const MessagesFieldDeleted = "deleted"

const UIDsTableName = "ui_ds"
const UIDsFieldUID = "uid"
const UIDsFieldDeleted = "deleted"
const UIDsFieldRecent = "recent"
const UIDsFieldMailboxID = "mailbox_ui_ds"
const UIDsFieldMessageID = "uid_message"

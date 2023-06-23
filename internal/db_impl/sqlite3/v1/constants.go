package v1

const DeletedSubscriptionsTableName = "deleted_subscriptions"
const DeletedSubscriptionsFieldName = "name"
const DeletedSubscriptionsFieldRemoteID = "remote_id"

const MailboxAttrsTableName = "mailbox_attrs_v2"
const MailboxAttrsFieldValue = "value"
const MailboxAttrsFieldMailboxID = "mailbox_id"

const MailboxFlagsTableName = "mailbox_flags_v2"
const MailboxFlagsFieldValue = "value"
const MailboxFlagsFieldMailboxID = "mailbox_id"

const MailboxPermFlagsTableName = "mailbox_perm_flags_v2"
const MailboxPermFlagsFieldValue = "value"
const MailboxPermFlagsFieldMailboxID = "mailbox_id"

const MailboxesTableName = "mailboxes_v2"
const MailboxesFieldID = "id"
const MailboxesFieldRemoteID = "remote_id"
const MailboxesFieldName = "name"
const MailboxesFieldUIDValidity = "uid_validity"
const MailboxesFieldSubscribed = "subscribed"

const MessageFlagsTableName = "message_flags_v2"
const MessageFlagsFieldValue = "value"
const MessageFlagsFieldMessageID = "message_id"

const MessagesTableName = "messages_v2"
const MessagesFieldID = "id"
const MessagesFieldRemoteID = "remote_id"
const MessagesFieldDate = "date"
const MessagesFieldSize = "size"
const MessagesFieldBody = "body"
const MessagesFieldBodyStructure = "body_structure"
const MessagesFieldEnvelope = "envelope"
const MessagesFieldDeleted = "deleted"

const MailboxMessagesFieldUID = "uid"
const MailboxMessagesFieldDeleted = "deleted"
const MailboxMessagesFieldRecent = "recent"
const MailboxMessagesFieldMailboxID = "mailbox_id"
const MailboxMessagesFieldMessageID = "message_id"
const MailboxMessagesFieldMessageRemoteID = "message_remote_id"

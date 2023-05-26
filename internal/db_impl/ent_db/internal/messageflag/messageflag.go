// Code generated by ent, DO NOT EDIT.

package messageflag

const (
	// Label holds the string label denoting the messageflag type in the database.
	Label = "message_flag"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldValue holds the string denoting the value field in the database.
	FieldValue = "value"
	// EdgeMessages holds the string denoting the messages edge name in mutations.
	EdgeMessages = "messages"
	// Table holds the table name of the messageflag in the database.
	Table = "message_flags"
	// MessagesTable is the table that holds the messages relation/edge.
	MessagesTable = "message_flags"
	// MessagesInverseTable is the table name for the Message entity.
	// It exists in this package in order to avoid circular dependency with the "message" package.
	MessagesInverseTable = "messages"
	// MessagesColumn is the table column denoting the messages relation/edge.
	MessagesColumn = "message_flags"
)

// Columns holds all SQL columns for messageflag fields.
var Columns = []string{
	FieldID,
	FieldValue,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "message_flags"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"message_flags",
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	for i := range ForeignKeys {
		if column == ForeignKeys[i] {
			return true
		}
	}
	return false
}
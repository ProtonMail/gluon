// Code generated by entc, DO NOT EDIT.

package messageflag

const (
	// Label holds the string label denoting the messageflag type in the database.
	Label = "message_flag"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldValue holds the string denoting the value field in the database.
	FieldValue = "value"
	// Table holds the table name of the messageflag in the database.
	Table = "message_flags"
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

package backend

// InternalIDKey is the key of the header entry we add to messages in the mailserver system.
// This allows us to detect when clients try to create a duplicate of a message, which we treat instead as a copy.
const InternalIDKey = `X-PM-GOMSRV-ID`

// InternalIDHeaderLength is the expected length of the full header entry, excluding the new line character.
const InternalIDHeaderLength = 52

// InternalIDHeaderLengthWithNewLine is the same as InternalIDHeaderLength, but includes the \r\n character.
const InternalIDHeaderLengthWithNewLine = InternalIDHeaderLength + 2

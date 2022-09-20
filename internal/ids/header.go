package ids

// InternalIDKey is the key of the header entry we add to messages in the mailserver system.
// This allows us to detect when clients try to create a duplicate of a message, which we treat instead as a copy.
const InternalIDKey = `X-Pm-Gluon-Id`

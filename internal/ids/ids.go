package ids

import (
	"fmt"
	"strings"

	"github.com/ProtonMail/gluon/imap"
)

const GluonRecoveryMailboxName = "Recovered Messages"
const GluonRecoveryMailboxNameLowerCase = "recovered messages"
const GluonInternalRecoveryMailboxRemoteID = imap.MailboxID("GLUON-INTERNAL-RECOVERY-MBOX")
const gluonInternalRecoveredMessageRemoteIDPrefix = "GLUON-RECOVERED-MESSAGE"

func NewRecoveredRemoteMessageID(internalID imap.InternalMessageID) imap.MessageID {
	return imap.MessageID(fmt.Sprintf("%v-%v", gluonInternalRecoveredMessageRemoteIDPrefix, internalID))
}

func IsRecoveredRemoteMessageID(id imap.MessageID) bool {
	return strings.HasPrefix(string(id), gluonInternalRecoveredMessageRemoteIDPrefix)
}

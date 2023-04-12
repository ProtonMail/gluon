package tests

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/internal/ids"
	goimap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestFetchAllMacro(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectAfternoonMeetingMailbox(t, client)

		newFetchCommand(t, client).withItems(goimap.FetchAll).fetch("1").forSeqNum(1, func(validator *validatorBuilder) {
			validator.wantFlags(goimap.RecentFlag)
			validator.wantEnvelope(newAfternoonMeetingMessageEnvelopeValidator)
			validator.wantSize(afternoonMeetingMessageDataSizeWithExtraHeader())
			// TODO: GOMSRV-175 - Timezone preservation.
			validator.wantInternalDate("08-Feb-1994 05:52:25 +0000")
		}).checkAndRequireMessageCount(1)
	})
}

func TestFetchFastMacro(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectAfternoonMeetingMailbox(t, client)

		newFetchCommand(t, client).withItems(goimap.FetchFast).fetch("1").forSeqNum(1, func(validator *validatorBuilder) {
			validator.wantFlags(goimap.RecentFlag)
			validator.wantSize(afternoonMeetingMessageDataSizeWithExtraHeader())
			// TODO: GOMSRV-175 - Timezone preservation.
			validator.wantInternalDate("08-Feb-1994 05:52:25 +0000")
		}).checkAndRequireMessageCount(1)
	})
}

func TestFetchFullMacro(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectAfternoonMeetingMailbox(t, client)

		newFetchCommand(t, client).withItems(goimap.FetchFull).fetch("1").forSeqNum(1, func(validator *validatorBuilder) {
			validator.wantFlags(goimap.RecentFlag)
			validator.wantEnvelope(newAfternoonMeetingMessageEnvelopeValidator)
			validator.wantSize(afternoonMeetingMessageDataSizeWithExtraHeader())
			validator.wantBodyStructure(validateAfternoonMeetingBodyStructure)
			// TODO: GOMSRV-175 - Timezone preservation.
			validator.wantInternalDate("08-Feb-1994 05:52:25 +0000")
		}).checkAndRequireMessageCount(1)
	})
}

func TestFetchRFC822(t *testing.T) {
	// This test does all checks for RFC822 at the same time so we can compare the retrieved data
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectAfternoonMeetingMailbox(t, client)

		// Load message
		fullMessageBytes, err := os.ReadFile("testdata/afternoon-meeting.eml")
		require.NoError(t, err)
		fullMessage := string(fullMessageBytes)

		newFetchCommand(t, client).withItems(goimap.FetchRFC822).fetch("1").forSeqNum(1, func(validator *validatorBuilder) {
			validator.ignoreFlags()
			validator.wantSectionString(goimap.FetchRFC822, func(t testing.TB, literal string) {
				messageFromSection := skipGLUONHeader(literal)
				require.Equal(t, fullMessage, messageFromSection)
			})
		}).checkAndRequireMessageCount(1)
	})
}

func TestFetchRFC822Header(t *testing.T) {
	// This test does all checks for RFC822 at the same time so we can compare the retrieved data
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectAfternoonMeetingMailbox(t, client)

		// Load message
		fullMessageBytes, err := os.ReadFile("testdata/afternoon-meeting.eml")
		require.NoError(t, err)
		fullMessage := string(fullMessageBytes)

		newFetchCommand(t, client).withItems(goimap.FetchRFC822Header).fetch("1").forSeqNum(1, func(validator *validatorBuilder) {
			validator.ignoreFlags()
			validator.wantSectionString(goimap.FetchRFC822Header, func(t testing.TB, literal string) {
				messageFromSection := skipGLUONHeader(literal)
				require.True(t, strings.HasPrefix(fullMessage, messageFromSection))
			})
		}).checkAndRequireMessageCount(1)
	})
}

func TestFetchRFC822Size(t *testing.T) {
	// This test does all checks for RFC822 at the same time so we can compare the retrieved data
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectAfternoonMeetingMailbox(t, client)
		newFetchCommand(t, client).withItems(goimap.FetchRFC822Size).fetch("1").forSeqNum(1, func(validator *validatorBuilder) {
			validator.ignoreFlags()
			validator.wantSize(afternoonMeetingMessageDataSizeWithExtraHeader())
		}).checkAndRequireMessageCount(1)
	})
}

func TestFetchRFC822Text(t *testing.T) {
	// This test does all checks for RFC822 at the same time so we can compare the retrieved data
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectAfternoonMeetingMailbox(t, client)

		// Load message
		fullMessageBytes, err := os.ReadFile("testdata/afternoon-meeting.eml")
		require.NoError(t, err)
		fullMessage := string(fullMessageBytes)

		newFetchCommand(t, client).withItems(goimap.FetchRFC822Text).fetch("1").forSeqNum(1, func(validator *validatorBuilder) {
			validator.ignoreFlags()
			validator.wantSectionString(goimap.FetchRFC822Text, func(t testing.TB, literal string) {
				messageFromSection := skipGLUONHeader(literal)
				require.True(t, strings.HasSuffix(fullMessage, messageFromSection))
			})
		}).checkAndRequireMessageCount(1)
	})
}

func TestFetchEnvelopeOnly(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectAfternoonMeetingMailbox(t, client)
		newFetchCommand(t, client).withItems(goimap.FetchEnvelope).fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.wantEnvelope(newAfternoonMeetingMessageEnvelopeValidator)
		}).checkAndRequireMessageCount(1)
	})
}

func TestFetchFlagsOnly(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectAfternoonMeetingMailbox(t, client)

		newFetchCommand(t, client).withItems(goimap.FetchFlags).fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.wantFlags(goimap.RecentFlag)
		}).checkAndRequireMessageCount(1)
	})
}

func TestFetchInternalDateOnly(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectAfternoonMeetingMailbox(t, client)

		newFetchCommand(t, client).withItems(goimap.FetchInternalDate).fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			// TODO: GOMSRV-175 - Timezone preservation.
			builder.wantInternalDate("08-Feb-1994 05:52:25 +0000")
		}).checkAndRequireMessageCount(1)
	})
}

func TestFetchUIDOnly(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectAfternoonMeetingMailbox(t, client)

		newFetchCommand(t, client).withItems(goimap.FetchUid).fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.wantUID(1)
		}).checkAndRequireMessageCount(1)
	})
}

func TestFetchRFC822AddsSeenFlag(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectAfternoonMeetingMailbox(t, client)

		// Fetch the message flags and check the seen flag is not set
		newFetchCommand(t, client).withItems(goimap.FetchFlags).fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.wantFlags(goimap.RecentFlag)
		}).check()

		// After Reading the body the seen flag should be set
		newFetchCommand(t, client).withItems(goimap.FetchRFC822).fetch("1").checkAndRequireMessageCount(1)

		// Fetch the message flags and check the seen flag is set
		newFetchCommand(t, client).withItems(goimap.FetchFlags).fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.wantFlags(goimap.RecentFlag, goimap.SeenFlag)
		}).check()
	})
}

func TestFetchRFC822TextAddsSeenFlag(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectAfternoonMeetingMailbox(t, client)

		// Fetch the message flags and check the seen flag is not set
		newFetchCommand(t, client).withItems(goimap.FetchFlags).fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.wantFlags(goimap.RecentFlag)
		}).check()

		// After Reading the body the Seen flag should be reported
		newFetchCommand(t, client).withItems(goimap.FetchRFC822Text).fetch("1").checkAndRequireMessageCount(1)

		// Fetch the message flags and check the seen flag is set
		newFetchCommand(t, client).withItems(goimap.FetchFlags).fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.wantFlags(goimap.RecentFlag, goimap.SeenFlag)
		}).check()
	})
}

func TestFetchRFC822HeaderDoesNotAddSeenFlag(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectAfternoonMeetingMailbox(t, client)

		// Fetch the message flags and check the seen flag is not set
		newFetchCommand(t, client).withItems(goimap.FetchFlags).fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.wantFlags(goimap.RecentFlag)
		}).check()

		// No flags should be reported when performing this fetch
		newFetchCommand(t, client).withItems(goimap.FetchRFC822Header).fetch("1").checkAndRequireMessageCount(1)

		// Fetch the message flags and check the seen flag is not set
		newFetchCommand(t, client).withItems(goimap.FetchFlags).fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.wantFlags(goimap.RecentFlag)
		}).check()
	})
}

func TestFetchBodyPeekDoesNotAddSeenFlag(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectAfternoonMeetingMailbox(t, client)

		// Fetch the message flags and check the seen flag is not set
		newFetchCommand(t, client).withItems(goimap.FetchFlags).fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.wantFlags(goimap.RecentFlag)
		}).check()

		// No flags should be reported when performing this fetch
		newFetchCommand(t, client).withItems("BODY.PEEK[TEXT]").fetch("1").checkAndRequireMessageCount(1)

		// Fetch the message flags and check the seen flag is not set
		newFetchCommand(t, client).withItems(goimap.FetchFlags).fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.wantFlags(goimap.RecentFlag)
		}).check()
	})
}

func TestFetchSequence(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectMailboxWithMultipleEntries(t, client)

		// delete message number 4
		require.NoError(t, client.Store(createSeqSet("4"), goimap.AddFlags, []interface{}{goimap.DeletedFlag}, nil))
		require.NoError(t, client.Expunge(nil))

		fetchResult := newFetchCommand(t, client).withItems(goimap.FetchEnvelope, goimap.FetchUid).fetch("4,1:2")
		fetchResult.forSeqNum(1, func(builder *validatorBuilder) {
			builder.wantUID(1)
			builder.wantEnvelope(func(builder *envelopeValidatorBuilder) {
				builder.wantTo("1@pm.me")
			})
		}).
			forSeqNum(2, func(builder *validatorBuilder) {
				builder.wantUID(2)
				builder.wantEnvelope(func(builder *envelopeValidatorBuilder) {
					builder.wantTo("2@pm.me")
				})
			}).
			forSeqNum(4, func(builder *validatorBuilder) {
				builder.wantUID(5)
				builder.wantEnvelope(func(builder *envelopeValidatorBuilder) {
					builder.wantTo("5@pm.me")
				})
			}).checkAndRequireMessageCount(3)

		// Results should be returned in ascending order
		require.Equal(t, uint32(1), fetchResult.messages[0].SeqNum)
		require.Equal(t, uint32(2), fetchResult.messages[1].SeqNum)
		require.Equal(t, uint32(4), fetchResult.messages[2].SeqNum)
	})
}

func TestFetchUID(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectMailboxWithMultipleEntries(t, client)

		// delete message number 4
		require.NoError(t, client.Store(createSeqSet("4"), goimap.AddFlags, []interface{}{goimap.DeletedFlag}, nil))
		require.NoError(t, client.Expunge(nil))

		messages := uidFetchMessagesClient(t, client, createSeqSet("2:5"), []goimap.FetchItem{goimap.FetchEnvelope, goimap.FetchUid})
		require.Equal(t, 3, len(messages))

		fetchResult := newFetchCommand(t, client).withItems(goimap.FetchEnvelope, goimap.FetchUid).fetchUid("5,2:4")
		fetchResult.forSeqNum(2, func(builder *validatorBuilder) {
			builder.wantUID(2)
			builder.wantEnvelope(func(builder *envelopeValidatorBuilder) {
				builder.wantTo("2@pm.me")
			})
		}).
			forSeqNum(3, func(builder *validatorBuilder) {
				builder.wantUID(3)
				builder.wantEnvelope(func(builder *envelopeValidatorBuilder) {
					builder.wantTo("3@pm.me")
				})
			}).
			forSeqNum(4, func(builder *validatorBuilder) {
				builder.wantUID(5)
				builder.wantEnvelope(func(builder *envelopeValidatorBuilder) {
					builder.wantTo("5@pm.me")
				})
			}).checkAndRequireMessageCount(3)

		// Results should be returned in ascending order
		// Results should be returned in ascending order
		require.Equal(t, uint32(2), fetchResult.messages[0].Uid)
		require.Equal(t, uint32(3), fetchResult.messages[1].Uid)
		require.Equal(t, uint32(5), fetchResult.messages[2].Uid)
	})
}

func TestFetchFromDataSequences(t *testing.T) {
	runOneToOneTestClientWithData(t, defaultServerOptions(t), func(client *client.Client, _ *testSession, _ string, _ imap.MailboxID) {
		const sectionStr = "BODY[HEADER.FIELDS (To Subject)]"
		fetchResult := newFetchCommand(t, client).withItems(sectionStr).fetch("1:4,30:31,81")
		fetchResult.forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection(sectionStr,
				`To: dovecot@procontrol.fi`,
				`Subject: [dovecot] first test mail`,
				``,
				``,
			)
		})
		fetchResult.forSeqNum(2, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection(sectionStr,
				`To: dovecot@procontrol.fi`,
				`Subject: [dovecot] Dovecot 0.93 released`,
				``,
				``,
			)
		})
		fetchResult.forSeqNum(3, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection(sectionStr,
				`To: dovecot@procontrol.fi`,
				`Subject: [dovecot] v0.95 released`,
				``,
				``,
			)
		})
		fetchResult.forSeqNum(4, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection(sectionStr,
				`To: "dovecot@procontrol.fi" <dovecot@procontrol.fi>`,
				`Subject: [dovecot] DOVECOT.PROCONTROL.FI`,
				``,
				``,
			)
		})
		fetchResult.forSeqNum(30, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection(sectionStr,
				`Subject: [dovecot] 0.98 released`,
				`To: dovecot@procontrol.fi`,
				``,
				``,
			)
		})
		fetchResult.forSeqNum(31, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection(sectionStr,
				`Subject: [dovecot] 0.98.1 released`,
				`To: dovecot@procontrol.fi`,
				``,
				``,
			)
		})
		fetchResult.forSeqNum(81, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection(sectionStr,
				`Subject: [dovecot] Re: Architectural questions`,
				`To: dovecot@procontrol.fi`,
				``,
				``,
			)
		})
		fetchResult.checkAndRequireMessageCount(7)
	})
}

func TestFetchFromDataUids(t *testing.T) {
	runOneToOneTestClientWithData(t, defaultServerOptions(t), func(client *client.Client, _ *testSession, _ string, _ imap.MailboxID) {
		// Remove a couple of messages
		require.NoError(t, client.Store(createSeqSet("20:29,50:60,90"), goimap.AddFlags, []interface{}{goimap.DeletedFlag}, nil))
		require.NoError(t, client.Expunge(nil))
		const sectionStr = "BODY[HEADER.FIELDS (To Subject)]"
		fetchResult := newFetchCommand(t, client).withItems(sectionStr).fetchUid("1:4,30:31,81")
		fetchResult.forUid(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection(sectionStr,
				`To: dovecot@procontrol.fi`,
				`Subject: [dovecot] first test mail`,
				``,
				``,
			)
		})
		fetchResult.forUid(2, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection(sectionStr,
				`To: dovecot@procontrol.fi`,
				`Subject: [dovecot] Dovecot 0.93 released`,
				``,
				``,
			)
		})
		fetchResult.forUid(3, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection(sectionStr,
				`To: dovecot@procontrol.fi`,
				`Subject: [dovecot] v0.95 released`,
				``,
				``,
			)
		})
		fetchResult.forUid(4, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection(sectionStr,
				`To: "dovecot@procontrol.fi" <dovecot@procontrol.fi>`,
				`Subject: [dovecot] DOVECOT.PROCONTROL.FI`,
				``,
				``,
			)
		})
		fetchResult.forUid(30, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection(sectionStr,
				`Subject: [dovecot] 0.98 released`,
				`To: dovecot@procontrol.fi`,
				``,
				``,
			)
		})
		fetchResult.forUid(31, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection(sectionStr,
				`Subject: [dovecot] 0.98.1 released`,
				`To: dovecot@procontrol.fi`,
				``,
				``,
			)
		})
		fetchResult.forUid(81, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection(sectionStr,
				`Subject: [dovecot] Re: Architectural questions`,
				`To: dovecot@procontrol.fi`,
				``,
				``,
			)
		})
		fetchResult.checkAndRequireMessageCount(7)
	})
}

func TestFetchInReplyTo(t *testing.T) {
	const message = `Received: with ECARTIS (v1.0.0; list dovecot); Tue, 23 Jul 2002 19:39:23 +0300 (EEST)
Return-Path: <cras@irccrew.org>
Delivered-To: dovecot@procontrol.fi
Date: Tue, 23 Jul 2002 19:39:23 +0300
From: Timo Sirainen <tss@iki.fi>
To: dovecot@procontrol.fi
Subject: [dovecot] first test mail
Message-ID: <20020723193923.J22431@irccrew.org>
User-Agent: Mutt/1.2.5i
Content-Type: text/plain; charset=us-ascii
Sender: dovecot-bounce@procontrol.fi
In-Reply-To: <Pine.LNX.4.44.0304061811480.10634-100000@allspice.nssg.mitel.com>

lets see if it works

`

	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		require.NoError(t, doAppendWithClient(client, "INBOX", message, time.Now()))
		_, err := client.Select("INBOX", false)
		require.NoError(t, err)
		newFetchCommand(t, client).withItems("ENVELOPE").
			fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantEnvelope(func(builder *envelopeValidatorBuilder) {
				builder.wantDateTime("23-Jul-2002 19:39:23 +0300")
				builder.wantSender("dovecot-bounce@procontrol.fi")
				builder.wantTo("dovecot@procontrol.fi")
				builder.wantFrom("tss@iki.fi")
				builder.wantSubject("[dovecot] first test mail")
				builder.wantMessageId("<20020723193923.J22431@irccrew.org>")
				builder.wantInReplyTo("<Pine.LNX.4.44.0304061811480.10634-100000@allspice.nssg.mitel.com>")
			})
		}).check()
	})
}

// --- helpers -------------------------------------------------------------------------------------------------------

func fillAndSelectAfternoonMeetingMailbox(t *testing.T, client *client.Client) {
	messageTime, err := time.Parse(goimap.DateTimeLayout, "07-Feb-1994 21:52:25 -0800")
	require.NoError(t, err)
	err = doAppendWithClientFromFile(t, client, "INBOX", "testdata/afternoon-meeting.eml", messageTime)
	require.NoError(t, err)
	_, err = client.Select("INBOX", false)
	require.NoError(t, err)
}

func newAfternoonMeetingMessageEnvelopeValidator(builder *envelopeValidatorBuilder) {
	fromAddress := &goimap.Address{
		PersonalName: "Fred Foobar",
		AtDomainList: "",
		MailboxName:  "foobar",
		HostName:     "Blurdybloop.COM",
	}
	toAddress := &goimap.Address{
		PersonalName: "",
		AtDomainList: "",
		MailboxName:  "mooch",
		HostName:     "owatagu.siam.edu",
	}

	builder.wantSubject("afternoon meeting")
	builder.wantAddressTypeFrom(fromAddress)
	builder.wantAddressTypeSender(fromAddress)
	builder.wantAddressTypeTo(toAddress)
	builder.wantDateTime("07-Feb-1994 21:52:25 -0800")
}

func validateAfternoonMeetingBodyStructure(builder *bodyStructureValidatorBuilder) {
	builder.wantMIMEType("text")
	builder.wantMIMESubType("plain")
	builder.wantLines(2)
	builder.wantSize(57)
	builder.wantParams(map[string]string{
		"charset": "US-ASCII",
	})
	builder.wantExtended(false)
}

func afternoonMeetingMessageDataSize() uint32 {
	fi, err := os.Stat("testdata/afternoon-meeting.eml")
	if err != nil {
		panic(err)
	}

	return uint32(fi.Size())
}

func afternoonMeetingMessageDataSizeWithExtraHeader() uint32 {
	return afternoonMeetingMessageDataSize() + uint32(len(ids.InternalIDKey)) + uint32(len(uuid.NewString())+4)
}

func fillAndSelectMailboxWithMultipleEntries(t *testing.T, client *client.Client) {
	const messageBoxName = "INBOX"

	require.NoError(t, client.Append(messageBoxName, []string{goimap.SeenFlag}, time.Now(), strings.NewReader("To: 1@pm.me")))
	require.NoError(t, client.Append(messageBoxName, nil, time.Now(), strings.NewReader("To: 2@pm.me")))
	require.NoError(t, client.Append(messageBoxName, []string{goimap.SeenFlag}, time.Now(), strings.NewReader("To: 3@pm.me")))
	require.NoError(t, client.Append(messageBoxName, nil, time.Now(), strings.NewReader("To: 4@pm.me")))
	require.NoError(t, client.Append(messageBoxName, []string{goimap.SeenFlag}, time.Now(), strings.NewReader("To: 5@pm.me")))
	_, err := client.Select(messageBoxName, false)
	require.NoError(t, err)
}

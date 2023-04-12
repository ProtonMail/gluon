package tests

import (
	"os"
	"testing"
	"time"

	goimap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/stretchr/testify/require"
)

func TestFetchBodySetsSeenFlag(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.doAppendFromFile(`INBOX`, `testdata/multipart-mixed.eml`).expect("OK")

		c.C(`A004 SELECT INBOX`)
		c.Se(`A004 OK [READ-WRITE] SELECT`)

		// The message initially has no flags except the recent flag.
		c.C(`A005 FETCH 1 (FLAGS)`)
		c.S(`* 1 FETCH (FLAGS (\Recent))`)
		c.OK("A005")

		// Fetch part of the body; the Seen flag should be implicitly set and included in the response.
		c.C(`A005 FETCH 1 (BODY[1.1])`)
		c.S(lines(`* 1 FETCH (BODY[1.1] {25}`,
			`*this */is**/_html_`,
			`**`,
			` FLAGS (\Recent \Seen))`,
		))
		c.OK("A005")

		// The message now has the seen flag.
		c.C(`A005 FETCH 1 (FLAGS)`)
		c.S(`* 1 FETCH (FLAGS (\Recent \Seen))`)
		c.OK("A005")
	})
}

func TestFetchBodyPeekDoesNotSetSeenFlag(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.doAppendFromFile(`INBOX`, `testdata/multipart-mixed.eml`).expect("OK")

		c.C(`A004 SELECT INBOX`)
		c.Se(`A004 OK [READ-WRITE] SELECT`)

		// The message initially has no flags other than the recent flag.
		c.C(`A005 FETCH 1 (FLAGS)`)
		c.S(`* 1 FETCH (FLAGS (\Recent))`)
		c.Sx(`A005 OK command completed in .*`)

		// Fetch part of the body via BODY.PEEK; the Seen flag should NOT be implicitly set.
		c.C(`A005 FETCH 1 (BODY.PEEK[1.1])`)
		c.S(lines(`* 1 FETCH (BODY[1.1] {25}`,
			`*this */is**/_html_`,
			`**`,
			`)`,
		))
		c.Sx(`A005 OK command completed in .*`)

		// The message still has no flags other than recent.
		c.C(`A005 FETCH 1 (FLAGS)`)
		c.S(`* 1 FETCH (FLAGS (\Recent))`)
		c.Sx(`A005 OK command completed in .*`)
	})
}

func TestFetchStructure(t *testing.T) {
	runOneToOneTestWithAuth(t, defaultServerOptions(t), func(c *testConnection, _ *testSession) {
		c.doAppendFromFile(`INBOX`, `testdata/multipart-mixed.eml`, `\Seen`).expect("OK")

		c.C(`A004 SELECT INBOX`)
		c.Se(`A004 OK [READ-WRITE] SELECT`)

		c.C(`A005 FETCH 1 (BODY)`)
		c.S(`* 1 FETCH (BODY ((("text" "plain" ("charset" "utf-8" "format" "flowed") NIL NIL "7bit" 25 2)("text" "html" ("charset" "utf-8") NIL NIL "7bit" 197 10) "alternative")("text" "plain" ("charset" "UTF-8" "name" "thing.txt" "x-mac-creator" "0" "x-mac-type" "0") NIL NIL "base64" 32 1) "mixed"))`)
		c.Sx(`A005 OK command completed in .*`)

		c.C(`A005 FETCH 1 (BODYSTRUCTURE)`)
		c.S(`* 1 FETCH (BODYSTRUCTURE ((("text" "plain" ("charset" "utf-8" "format" "flowed") NIL NIL "7bit" 25 2 NIL NIL NIL NIL)("text" "html" ("charset" "utf-8") NIL NIL "7bit" 197 10 NIL NIL NIL NIL) "alternative" ("boundary" "------------62DCF50B21CF279F489F0184") NIL NIL NIL)("text" "plain" ("charset" "UTF-8" "name" "thing.txt" "x-mac-creator" "0" "x-mac-type" "0") NIL NIL "base64" 32 1 NIL ("attachment" ("filename" "thing.txt")) NIL NIL) "mixed" ("boundary" "------------4AC5F36D876D5EED478B5FF9") NIL "en-US" NIL))`)
		c.Sx(`A005 OK command completed in .*`)
	})
}

func TestFetchStructureMultiPart(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectMultiPartMessage(t, client)

		newFetchCommand(t, client).withItems(goimap.FetchBodyStructure).fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.wantBodyStructure(func(builder *bodyStructureValidatorBuilder) {
				builder.wantMIMEType("multipart")
				builder.wantMIMESubType("mixed")
				builder.wantLanguage("en-US")
				builder.wantParams(map[string]string{"boundary": "------------4AC5F36D876D5EED478B5FF9"})
				builder.wantPart(func(builder *bodyStructureValidatorBuilder) {
					builder.wantMIMEType("multipart")
					builder.wantMIMESubType("alternative")
					builder.wantParams(map[string]string{"boundary": "------------62DCF50B21CF279F489F0184"})
					builder.wantPart(func(builder *bodyStructureValidatorBuilder) {
						builder.wantMIMEType("text")
						builder.wantMIMESubType("plain")
						builder.wantSize(25)
						builder.wantEncoding("7bit")
						builder.wantParams(map[string]string{
							"charset": "utf-8",
							"format":  "flowed",
						})
						builder.wantLines(2)
					})
					builder.wantPart(func(builder *bodyStructureValidatorBuilder) {
						builder.wantMIMEType("text")
						builder.wantMIMESubType("html")
						builder.wantSize(197)
						builder.wantEncoding("7bit")
						builder.wantParams(map[string]string{
							"charset": "utf-8",
						})
						builder.wantLines(10)
					})
				})
				builder.wantPart(func(builder *bodyStructureValidatorBuilder) {
					builder.wantMIMEType("text")
					builder.wantMIMESubType("plain")
					builder.wantSize(32)
					builder.wantEncoding("base64")
					builder.wantParams(map[string]string{
						"charset":       "UTF-8",
						"name":          "thing.txt",
						"x-mac-creator": "0",
						"x-mac-type":    "0",
					})
					builder.wantDisposition("attachment")
					builder.wantDispositionParams(map[string]string{
						"filename": "thing.txt",
					})
					builder.wantLines(1)
				})
			})
		}).checkAndRequireMessageCount(1)
	})
}

func TestFetchEnvelopeMultiPart(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectMultiPartMessage(t, client)

		newFetchCommand(t, client).withItems(goimap.FetchEnvelope).fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.wantEnvelope(func(builder *envelopeValidatorBuilder) {
				address := &goimap.Address{
					PersonalName: "BQA",
					AtDomainList: "",
					MailboxName:  "somebody",
					HostName:     "gmail.com",
				}
				addressTo := &goimap.Address{
					PersonalName: "",
					AtDomainList: "",
					MailboxName:  "somebody",
					HostName:     "gmail.com",
				}
				builder.wantAddressTypeTo(addressTo)
				builder.wantAddressTypeSender(address)
				builder.wantAddressTypeFrom(address)
				builder.wantAddressTypeReplyTo(addressTo)
				builder.wantDateTime("26-Mar-2021 20:01:23 +0100")
				builder.wantSubject("Simple test mail")
			})
		}).checkAndRequireMessageCount(1)
	})
}

func TestFetchStructureEmbedded(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectEmbeddedMessage(t, client)

		newFetchCommand(t, client).withItems(goimap.FetchBodyStructure).fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.wantBodyStructure(func(builder *bodyStructureValidatorBuilder) {
				builder.wantMIMEType("multipart")
				builder.wantMIMESubType("mixed")
				builder.wantParams(map[string]string{"boundary": "simple boundary"})
				builder.wantPart(func(builder *bodyStructureValidatorBuilder) {
					builder.wantMIMEType("text")
					builder.wantMIMESubType("plain")
					builder.wantSize(40)
					builder.wantParams(map[string]string{
						"charset": "us-ascii",
					})
					builder.wantSize(40)
					builder.wantLines(1)
				})
				builder.wantPart(func(builder *bodyStructureValidatorBuilder) {
					builder.wantMIMEType("multipart")
					builder.wantMIMESubType("rfc822")
					builder.wantParams(map[string]string{
						"name": "test.eml",
					})
					builder.wantDispositionParams(map[string]string{
						"filename": "test.eml",
					})
					builder.wantDisposition("attachment")
					builder.wantPart(func(builder *bodyStructureValidatorBuilder) {
						builder.wantMIMEType("text")
						builder.wantMIMESubType("plain")
						builder.wantSize(36)
						builder.wantParams(map[string]string{
							"charset": "us-ascii",
						})
						builder.wantLines(4)
					})
					builder.wantPart(func(builder *bodyStructureValidatorBuilder) {
						builder.wantMIMEType("text")
						builder.wantMIMESubType("plain")
						builder.wantSize(26)
						builder.wantParams(map[string]string{
							"charset": "us-ascii",
						})
						builder.wantLines(1)
					})
				})
			})
		}).checkAndRequireMessageCount(1)
	})
}

func TestFetchBodyMultiPart(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectMultiPartMessage(t, client)

		// Get full body
		newFetchCommand(t, client).withItems("BODY[]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			fullMessageBytes, err := os.ReadFile("testdata/multipart-mixed.eml")
			require.NoError(t, err)
			fullMessage := string(fullMessageBytes)
			builder.ignoreFlags()
			builder.wantSectionAndSkipGLUONHeader("BODY[]", fullMessage)
		}).check()

		// Get first message part
		newFetchCommand(t, client).withItems("BODY[1.TEXT]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection("BODY[1.TEXT]",
				``,
				`--------------62DCF50B21CF279F489F0184`,
				`Content-Type: text/plain; charset=utf-8; format=flowed`,
				`Content-Transfer-Encoding: 7bit`,
				``,
				`*this */is**/_html_`,
				`**`,
				``,
				`--------------62DCF50B21CF279F489F0184`,
				`Content-Type: text/html; charset=utf-8`,
				`Content-Transfer-Encoding: 7bit`,
				``,
				`<html>`,
				`  <head>`,
				``,
				`    <meta http-equiv="content-type" content="text/html; charset=UTF-8">`,
				`  </head>`,
				`  <body>`,
				`    <b>this </b><i>is<b> </b></i><u>html</u><br>`,
				`    <b></b>`,
				`  </body>`,
				`</html>`,
				``,
				`--------------62DCF50B21CF279F489F0184--`,
				``,
			)
		}).check()

		// Get first subpart from first part
		newFetchCommand(t, client).withItems("BODY[1.1.TEXT]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection("BODY[1.1.TEXT]",
				`*this */is**/_html_`,
				`**`,
				``,
			)
		}).check()

		// Get second subpart from first part
		newFetchCommand(t, client).withItems("BODY[1.2.TEXT]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection("BODY[1.2.TEXT]", `<html>`,
				`  <head>`,
				``,
				`    <meta http-equiv="content-type" content="text/html; charset=UTF-8">`,
				`  </head>`,
				`  <body>`,
				`    <b>this </b><i>is<b> </b></i><u>html</u><br>`,
				`    <b></b>`,
				`  </body>`,
				`</html>`,
				``,
			)
		}).check()

		// Get second message part
		newFetchCommand(t, client).withItems("BODY[2.TEXT]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection("BODY[2.TEXT]", "dGhpcyBpcyBteSBhdHRhY2htZW50Cg==")
		}).check()
	})
}

func TestFetchBodyEmbedded(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectEmbeddedMessage(t, client)

		// Get full body
		newFetchCommand(t, client).withItems("BODY[]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			fullMessageBytes, err := os.ReadFile("testdata/embedded-rfc822.eml")
			require.NoError(t, err)
			fullMessage := string(fullMessageBytes)
			builder.ignoreFlags()
			builder.wantSectionAndSkipGLUONHeader("BODY[]", fullMessage)
		}).check()

		// Get first message part
		newFetchCommand(t, client).withItems("BODY[1.TEXT]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection("BODY[1.TEXT]",
				`This part does not end with a linebreak.`,
			)
		}).check()

		// Get second message part
		newFetchCommand(t, client).withItems("BODY[2.TEXT]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection("BODY[2.TEXT]",
				`--embedded-boundary`,
				`Content-type: text/plain; charset=us-ascii`,
				``,
				`This part is embedded`,
				``,
				`--`,
				`From me`,
				`--embedded-boundary`,
				`Content-type: text/plain; charset=us-ascii`,
				``,
				`This part is also embedded`,
				`--embedded-boundary--`,
				``,
			)
		}).check()

		// fetch first subpart of the second message
		newFetchCommand(t, client).withItems("BODY[2.1.TEXT]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection("BODY[2.1.TEXT]", `This part is embedded`,
				``,
				`--`,
				`From me`,
			)
		}).check()

		// fetch second subpart of the second message
		newFetchCommand(t, client).withItems("BODY[2.2.TEXT]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection("BODY[2.2.TEXT]", `This part is also embedded`)
		}).check()
	})
}

func TestFetchBodyPlain(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectPlainMessage(t, client)

		// Get full body
		newFetchCommand(t, client).withItems("BODY[]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			fullMessageBytes, err := os.ReadFile("testdata/text-plain.eml")
			require.NoError(t, err)
			fullMessage := string(fullMessageBytes)
			builder.ignoreFlags()
			builder.wantSectionAndSkipGLUONHeader("BODY[]", fullMessage)
		}).check()

		// Get first message part
		newFetchCommand(t, client).withItems("BODY[TEXT]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection("BODY[TEXT]", `This is body of mail =F0=9F=91=8B`,
				``,
				``,
				``,
			)
		}).check()
	})
}

func TestFetchStructurePlain(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectPlainMessage(t, client)

		newFetchCommand(t, client).withItems(goimap.FetchBodyStructure).fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.wantBodyStructure(func(builder *bodyStructureValidatorBuilder) {
				builder.wantMIMEType("text")
				builder.wantMIMESubType("plain")
				builder.wantEncoding("quoted-printable")
				builder.wantSize(39)
				builder.wantLines(3)
				builder.wantParams(map[string]string{
					"charset": "utf-8",
				})
			})
		}).checkAndRequireMessageCount(1)
	})
}

func TestFetchBodyPartialMultiPart(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectMultiPartMessage(t, client)

		newFetchCommand(t, client).withItems("BODY[1.TEXT]<20.10>").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			messageText := "F50B21CF27"
			builder.ignoreFlags().wantSection("BODY[1.TEXT]<20>", messageText)
		}).check()

		// Get first subpart from first part
		newFetchCommand(t, client).withItems("BODY[1.1.TEXT]<8.3>").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			messageText := lines(
				`is*`,
			)
			builder.ignoreFlags()
			builder.wantSection("BODY[1.1.TEXT]<8>", messageText)
		}).check()
	})
}

func TestFetchBodyPartialReturnsEmptyWhenStartingOctetIsGreaterThanContentSize(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectMultiPartMessage(t, client)

		newFetchCommand(t, client).withItems("BODY[1.TEXT]<20000.10>").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSectionEmpty("BODY[1.TEXT]<20>")
		}).check()
	})
}

func TestFetchHeaderMultiPart(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectMultiPartMessage(t, client)

		newFetchCommand(t, client).withItems("BODY[HEADER]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSectionAndSkipGLUONHeader("BODY[HEADER]",
				`Return-Path: <somebody@gmail.com>`,
				`Received: from [10.1.1.121] ([185.159.157.131])`,
				`        by smtp.gmail.com with ESMTPSA id t8sm14889112wrr.10.2021.03.26.12.01.23`,
				`        for <somebody@gmail.com>`,
				`        (version=TLS1_3 cipher=TLS_AES_128_GCM_SHA256 bits=128/128);`,
				`        Fri, 26 Mar 2021 12:01:24 -0700 (PDT)`,
				`To: somebody@gmail.com`,
				`From: BQA <somebody@gmail.com>`,
				`Subject: Simple test mail`,
				`Message-ID: <d3b6e735-8fb4-9bde-1063-049b97f8f3ca@gmail.com>`,
				`Date: Fri, 26 Mar 2021 20:01:23 +0100`,
				`User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:78.0)`,
				` Gecko/20100101 Thunderbird/78.8.1`,
				`MIME-Version: 1.0`,
				`Content-Type: multipart/mixed;`,
				` boundary="------------4AC5F36D876D5EED478B5FF9"`,
				`Content-Language: en-US`,
				``,
				``,
			)
		}).check()

		newFetchCommand(t, client).withItems("BODY[1.HEADER]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection("BODY[1.HEADER]",
				`Content-Type: multipart/alternative;`,
				` boundary="------------62DCF50B21CF279F489F0184"`,
				``,
				``,
			)
		}).check()

		newFetchCommand(t, client).withItems("BODY[1.1.HEADER]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection("BODY[1.1.HEADER]",
				`Content-Type: text/plain; charset=utf-8; format=flowed`,
				`Content-Transfer-Encoding: 7bit`,
				``,
				``,
			)
		}).check()

		newFetchCommand(t, client).withItems("BODY[1.2.HEADER]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection("BODY[1.2.HEADER]",
				`Content-Type: text/html; charset=utf-8`,
				`Content-Transfer-Encoding: 7bit`,
				``,
				``,
			)
		}).check()

		// Get second message part
		newFetchCommand(t, client).withItems("BODY[2.HEADER]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection("BODY[2.HEADER]",
				`Content-Type: text/plain; charset=UTF-8; x-mac-type="0"; x-mac-creator="0";`,
				` name="thing.txt"`,
				`Content-Transfer-Encoding: base64`,
				`Content-Disposition: attachment;`,
				` filename="thing.txt"`,
				``,
				``,
			)
		}).check()
	})
}

func TestFetchHeaderFieldsMultiPart(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectMultiPartMessage(t, client)

		newFetchCommand(t, client).withItems("BODY[HEADER.FIELDS (To From Date)]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.wantSectionAndSkipGLUONHeader("BODY[HEADER.FIELDS (To From Date)]",
				`To: somebody@gmail.com`,
				`From: BQA <somebody@gmail.com>`,
				`Date: Fri, 26 Mar 2021 20:01:23 +0100`,
				``,
				``,
			)
		}).check()

		newFetchCommand(t, client).withItems("BODY[HEADER.FIELDS.NOT (To From Date)]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSectionAndSkipGLUONHeader("BODY[HEADER.FIELDS.NOT (To From Date)]",
				`Return-Path: <somebody@gmail.com>`,
				`Received: from [10.1.1.121] ([185.159.157.131])`,
				`        by smtp.gmail.com with ESMTPSA id t8sm14889112wrr.10.2021.03.26.12.01.23`,
				`        for <somebody@gmail.com>`,
				`        (version=TLS1_3 cipher=TLS_AES_128_GCM_SHA256 bits=128/128);`,
				`        Fri, 26 Mar 2021 12:01:24 -0700 (PDT)`,
				`Subject: Simple test mail`,
				`Message-ID: <d3b6e735-8fb4-9bde-1063-049b97f8f3ca@gmail.com>`,
				`User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:78.0)`,
				` Gecko/20100101 Thunderbird/78.8.1`,
				`MIME-Version: 1.0`,
				`Content-Type: multipart/mixed;`,
				` boundary="------------4AC5F36D876D5EED478B5FF9"`,
				`Content-Language: en-US`,
				``,
				``,
			)
		}).check()
	})
}

func TestFetchHeaderFieldsEmbedded(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectEmbeddedMessage(t, client)

		newFetchCommand(t, client).withItems("BODY[2.HEADER.FIELDS (To)]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSectionAndSkipGLUONHeader("BODY[2.HEADER.FIELDS (To)]",
				`To: someone`,
				``,
				``,
			)
		}).check()

		newFetchCommand(t, client).withItems("BODY[2.HEADER.FIELDS.NOT (To)]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSectionAndSkipGLUONHeader("BODY[2.HEADER.FIELDS.NOT (To)]",
				`Subject: Fwd: embedded`,
				`Content-type: multipart/mixed; boundary="embedded-boundary"`,
				``,
				``,
			)
		}).check()
	})
}

func TestFetchMIMEMultiPart(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectMultiPartMessage(t, client)

		newFetchCommand(t, client).withItems("BODY[MIME]").fetchFailure("1")

		newFetchCommand(t, client).withItems("BODY[1.MIME]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection("BODY[1.MIME]",
				`Content-Type: multipart/alternative;`,
				` boundary="------------62DCF50B21CF279F489F0184"`,
				``,
				``,
			)
		}).check()

		newFetchCommand(t, client).withItems("BODY[1.1.MIME]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection("BODY[1.1.MIME]",
				`Content-Type: text/plain; charset=utf-8; format=flowed`,
				`Content-Transfer-Encoding: 7bit`,
				``,
				``,
			)
		}).check()

		newFetchCommand(t, client).withItems("BODY[1.2.MIME]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection("BODY[1.2.MIME]",
				`Content-Type: text/html; charset=utf-8`,
				`Content-Transfer-Encoding: 7bit`,
				``,
				``,
			)
		}).check()

		// Get second message part
		newFetchCommand(t, client).withItems("BODY[2.MIME]").fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection("BODY[2.MIME]",
				`Content-Type: text/plain; charset=UTF-8; x-mac-type="0"; x-mac-creator="0";`,
				` name="thing.txt"`,
				`Content-Transfer-Encoding: base64`,
				`Content-Disposition: attachment;`,
				` filename="thing.txt"`,
				``,
				``,
			)
		}).check()
	})
}

func TestFetchBodyPeekSection1(t *testing.T) {
	const message = `Received: with ECARTIS (v1.0.0; list dovecot); Wed, 31 Jul 2002 22:48:41 +0300 (EEST)
Return-Path: <cras@irccrew.org>
Delivered-To: dovecot@procontrol.fi
Received: from shodan.irccrew.org (shodan.irccrew.org [80.83.4.2])
	by danu.procontrol.fi (Postfix) with ESMTP id F141123829
	for <dovecot@procontrol.fi>; Wed, 31 Jul 2002 22:48:40 +0300 (EEST)
Received: by shodan.irccrew.org (Postfix, from userid 6976)
	id 42ED44C0A0; Wed, 31 Jul 2002 22:48:40 +0300 (EEST)
Date: Wed, 31 Jul 2002 22:48:39 +0300
From: Timo Sirainen <tss@iki.fi>
To: dovecot@procontrol.fi
Subject: [dovecot] v0.95 released
Message-ID: <20020731224839.H22431@irccrew.org>
Mime-Version: 1.0
Content-Disposition: inline
User-Agent: Mutt/1.2.5i
Content-Type: text/plain; charset=us-ascii
X-archive-position: 3
X-ecartis-version: Ecartis v1.0.0
Sender: dovecot-bounce@procontrol.fi
Errors-to: dovecot-bounce@procontrol.fi
X-original-sender: tss@iki.fi
Precedence: bulk
X-list: dovecot
X-UID: 3                                                  
Status: O

v0.95 2002-07-31  Timo Sirainen <tss@iki.fi>

	+ Initial SSL support using GNU TLS, tested with v0.5.1.
	  TLS support is still missing.
	+ Digest-MD5 authentication method
	+ passwd-file authentication backend
	+ Code cleanups
	- Found several bugs from mempool and ioloop code, now we should
	  be stable? :)
	- A few corrections for long header field handling

`

	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		{
			require.NoError(t, doAppendWithClient(client, "INBOX", message, time.Now()))
			_, err := client.Select("INBOX", false)
			require.NoError(t, err)

			newFetchCommand(t, client).withItems("BODY.PEEK[1]").fetchUid("1").forUid(1, func(builder *validatorBuilder) {
				const expectedBody = `v0.95 2002-07-31  Timo Sirainen <tss@iki.fi>

	+ Initial SSL support using GNU TLS, tested with v0.5.1.
	  TLS support is still missing.
	+ Digest-MD5 authentication method
	+ passwd-file authentication backend
	+ Code cleanups
	- Found several bugs from mempool and ioloop code, now we should
	  be stable? :)
	- A few corrections for long header field handling

`
				builder.ignoreFlags()
				builder.wantSection("BODY[1]", expectedBody)
			}).check()
		}
	})
}

func TestFetchBodyPeekInvalidSectionNumber(t *testing.T) {
	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		fillAndSelectPlainMessage(t, client)
		newFetchCommand(t, client).withItems("BODY.PEEK[2]").fetchFailure("1")
	})
}

func TestFetchBody_Dovecot_InvalidMessageHeader(t *testing.T) {
	// This tests fails when requesting when fetch BODY.PEEK[HEADER.FIELDS (In-Reply-To In-Reply-To Cc)].
	// Instead of only returning In-Reply-To and Cc, it was also returning the References header.
	const message = `Received: with ECARTIS (v1.0.0; list dovecot); Mon, 07 Apr 2003 01:27:38 +0300 (EEST)
Return-Path: <tss@iki.fi>
X-Original-To: dovecot@procontrol.fi
Delivered-To: dovecot@procontrol.fi
Received: from oma.irssi.org (ip213-185-36-189.laajakaista.mtv3.fi [213.185.36.189])
	by danu.procontrol.fi (Postfix) with ESMTP id 2A7282387F
	for <dovecot@procontrol.fi>; Mon,  7 Apr 2003 01:27:38 +0300 (EEST)
Received: by oma.irssi.org (Postfix, from userid 1000)
	id 3D1565E01F95; Mon,  7 Apr 2003 01:27:36 +0300 (EEST)
Subject: [dovecot] Re: message order reversed on copying
From: Timo Sirainen <tss@iki.fi>
To: Charlie Brady <charlieb-dovecot@e-smith.com>
Cc: dovecot@procontrol.fi
In-Reply-To: <Pine.LNX.4.44.0304061811480.10634-100000@allspice.nssg.mitel.com>
References:
	 <Pine.LNX.4.44.0304061811480.10634-100000@allspice.nssg.mitel.com>
Content-Type: text/plain
Content-Transfer-Encoding: 7bit
Organization:
Message-Id: <1049668055.22903.405.camel@hurina>
Mime-Version: 1.0
X-Mailer: Ximian Evolution 1.2.3
Date: 07 Apr 2003 01:27:36 +0300
X-archive-position: 516
X-ecartis-version: Ecartis v1.0.0
Sender: dovecot-bounce@procontrol.fi
Errors-to: dovecot-bounce@procontrol.fi
X-original-sender: tss@iki.fi
Precedence: bulk
X-list: dovecot
X-UID: 516
Status: O

On Mon, 2003-04-07 at 01:15, Charlie Brady wrote:
> > > +  nfiles = scandir (tmp, &names, NULL, maildir_namesort);
> > scandir() isn't portable though.
> No? Which OS doesn't have it?

It's not in any standard -> it's not portable. Maybe it's portable
enough, but I try to avoid those whenever possible.

> > Which changes UIDVALIDITY and forces clients to discard their cache.
>
> Yes, I could have mentioned that. Still, it's the only way I know to
> restore the sorting order once it's jumbled.

How about making your client sort the messages by received-date? :)

`

	runOneToOneTestClientWithAuth(t, defaultServerOptions(t), func(client *client.Client, _ *testSession) {
		require.NoError(t, doAppendWithClient(client, "INBOX", message, time.Now()))
		_, err := client.Select("INBOX", false)
		require.NoError(t, err)
		newFetchCommand(t, client).withItems("BODY.PEEK[HEADER.FIELDS (In-Reply-To In-Reply-To Cc)]").
			fetch("1").forSeqNum(1, func(builder *validatorBuilder) {
			builder.ignoreFlags()
			builder.wantSection(`BODY[HEADER.FIELDS (In-Reply-To In-Reply-To Cc)]`,
				`Cc: dovecot@procontrol.fi
In-Reply-To: <Pine.LNX.4.44.0304061811480.10634-100000@allspice.nssg.mitel.com>

`)
		}).check()
	})
}

// --- helpers -------------------------------------------------------------------------------------------------------
// NOTE: these helpers create messages with the seen flag to avoid interfering with the go imap client library. Due to
// a timing issue, it is possible that a fetch update for the flag state can be received as a separate message.
// The current validation mechanism can't handle that, so we try to avoid it all together here.
func fillAndSelectMultiPartMessage(t *testing.T, client *client.Client) {
	messageTime, err := time.Parse(goimap.DateTimeLayout, "07-Feb-1994 21:52:25 -0800")
	require.NoError(t, err)
	err = doAppendWithClientFromFile(t, client, "INBOX", "testdata/multipart-mixed.eml", messageTime, goimap.SeenFlag)
	require.NoError(t, err)
	_, err = client.Select("INBOX", false)
	require.NoError(t, err)
}

func fillAndSelectEmbeddedMessage(t *testing.T, client *client.Client) {
	messageTime, err := time.Parse(goimap.DateTimeLayout, "07-Feb-1994 21:52:25 -0800")
	require.NoError(t, err)
	err = doAppendWithClientFromFile(t, client, "INBOX", "testdata/embedded-rfc822.eml", messageTime, goimap.SeenFlag)
	require.NoError(t, err)
	_, err = client.Select("INBOX", false)
	require.NoError(t, err)
}

func fillAndSelectPlainMessage(t *testing.T, client *client.Client) {
	messageTime, err := time.Parse(goimap.DateTimeLayout, "07-Feb-1994 21:52:25 -0800")
	require.NoError(t, err)
	err = doAppendWithClientFromFile(t, client, "INBOX", "testdata/text-plain.eml", messageTime, goimap.SeenFlag)
	require.NoError(t, err)
	_, err = client.Select("INBOX", false)
	require.NoError(t, err)
}

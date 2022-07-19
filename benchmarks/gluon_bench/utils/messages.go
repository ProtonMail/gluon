package utils

// Hardcoded messages used to generate mailboxes

const MessageMultiPartMixed = `Return-Path: <somebody@gmail.com>
Received: from [10.1.1.121] ([185.159.157.131])
        by smtp.gmail.com with ESMTPSA id t8sm14889112wrr.10.2021.03.26.12.01.23
        for <somebody@gmail.com>
        (version=TLS1_3 cipher=TLS_AES_128_GCM_SHA256 bits=128/128);
        Fri, 26 Mar 2021 12:01:24 -0700 (PDT)
To: somebody@gmail.com
From: BQA <somebody@gmail.com>
Subject: Simple test mail
Message-ID: <d3b6e735-8fb4-9bde-1063-049b97f8f3ca@gmail.com>
Date: Fri, 26 Mar 2021 20:01:23 +0100
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:78.0)
 Gecko/20100101 Thunderbird/78.8.1
MIME-Version: 1.0
Content-Type: multipart/mixed;
 boundary="------------4AC5F36D876D5EED478B5FF9"
Content-Language: en-US

This is a multi-part message in MIME format.
--------------4AC5F36D876D5EED478B5FF9
Content-Type: multipart/alternative;
 boundary="------------62DCF50B21CF279F489F0184"


--------------62DCF50B21CF279F489F0184
Content-Type: text/plain; charset=utf-8; format=flowed
Content-Transfer-Encoding: 7bit

*this */is**/_html_
**

--------------62DCF50B21CF279F489F0184
Content-Type: text/html; charset=utf-8
Content-Transfer-Encoding: 7bit

<html>
  <head>

    <meta http-equiv="content-type" content="text/html; charset=UTF-8">
  </head>
  <body>
    <b>this </b><i>is<b> </b></i><u>html</u><br>
    <b></b>
  </body>
</html>

--------------62DCF50B21CF279F489F0184--

--------------4AC5F36D876D5EED478B5FF9
Content-Type: text/plain; charset=UTF-8; x-mac-type="0"; x-mac-creator="0";
 name="thing.txt"
Content-Transfer-Encoding: base64
Content-Disposition: attachment;
 filename="thing.txt"

dGhpcyBpcyBteSBhdHRhY2htZW50Cg==
--------------4AC5F36D876D5EED478B5FF9--

`

const MessageAfterNoonMeeting = `Date: Mon, 7 Feb 1994 21:52:25 -0800 (PST)
From: Fred Foobar <foobar@Blurdybloop.COM>
Subject: afternoon meeting
To: mooch@owatagu.siam.edu
Message-Id: <B27397-0100000@Blurdybloop.COM>
MIME-Version: 1.0
Content-Type: TEXT/PLAIN; CHARSET=US-ASCII

Hello Joe, do you think we can meet at 3:30 tomorrow?
`

const MessageEmbedded = `From: Nathaniel Borenstein <nsb@bellcore.com>
To:  Ned Freed <ned@innosoft.com>
Subject: Sample message
MIME-Version: 1.0
Content-type: multipart/mixed; boundary="simple boundary"

This is the preamble.  It is to be ignored, though it
is a handy place for mail composers to include an
explanatory note to non-MIME compliant readers.
--simple boundary
Content-type: text/plain; charset=us-ascii

This part does not end with a linebreak.
--simple boundary
Content-Disposition: attachment; filename=test.eml
Content-Type: message/rfc822; name=test.eml
X-Pm-Content-Encryption: on-import

To: someone
Subject: Fwd: embedded
Content-type: multipart/mixed; boundary="embedded-boundary"

--embedded-boundary
Content-type: text/plain; charset=us-ascii

This part is embedded

--
From me
--embedded-boundary
Content-type: text/plain; charset=us-ascii

This part is also embedded
--embedded-boundary--

--simple boundary--
This is the epilogue.  It is also to be ignored.`

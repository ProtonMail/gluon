parser grammar IMAPParser;

options {
	tokenVocab = IMAPLexer;
}

// 2.2. Commands and Responses
crlf: CR LF;

// 2.2.1. Client Protocol Sender and Server Protocol Receiver
tag: tagChar+;

// 4.1. Data Formats
atom: atomChar+;

// 4.2. Number
number: digit+;

// 4.3. String
string
	: quoted 		# stringQtd
	| literal		# stringLit
	;

nstring
    : string
    | N I L
    ;

quoted: DQuote quotedChar* DQuote;

quotedChar
	: unescapedChar							# rawQChar
	| Backslash quotedSpecial		# escQChar
	;

quotedSpecial: DQuote | Backslash;

literal: LCurly number RCurly crlf uuid;
uuid: hex4 hex4 Minus hex4 Minus hex4 Minus hex4 Minus hex4 hex4 hex4;
hex4: hex hex hex hex;
hex:  alpha | digit;

astring
	: astringChar+ 	# astringRaw
	| string				# astringStr
	;

// 6. Client Commands
command: ((tag SP (commandAny | commandNonAuth | commandAuth | commandSelected)) | done) crlf EOF;

// 6.1. Client Commands - Any State
commandAny: capability | noop | logout | id;

// 6.1.1 CAPABILITY Command
capability: C A P A B I L I T Y;

// 6.1.2 NOOP Command
noop: N O O P;

// 6.1.3 LOGOUT Command
logout: L O G O U T;

// 6.2. Client Commands - Not Authenticated State
commandNonAuth: login | auth | startTLS;

// 6.2.1. STARTTLS Command
startTLS: S T A R T T L S;

// 6.2.2. AUTHENTICATE Command
auth:  A U T H E N T I C A T E SP authType (crlf base64)*;

authType: atom;

base64: (base64Char base64Char base64Char base64Char)* base64Terminal?;

base64Char: alpha | digit | Plus | Slash;

base64Terminal
	: base64Char base64Char Equal Equal
	| base64Char base64Char base64Char Equal
	;

// 6.2.3. LOGIN Command
login: L O G I N SP userID SP password;

userID: astring;

password: astring;

// 6.3. Client Commands - Authenticated State
commandAuth: append | create | del | examine | list | lsub | rename | select | status | sub | unsub | idle;

mailbox
	: I N B O X		# mboxInbox
	| astring			# mboxOther
	;

listMailbox
	: listChar+		# listMboxRaw
	| string			# listMboxStr
	;

listChar: atomChar | listWildcards | respSpecials;

listWildcards: Percent | Asterisk;

// 6.3.1. SELECT Command
select: S E L E C T SP mailbox;

// 6.3.2. EXAMINE Command
examine: E X A M I N E SP mailbox;

// 6.3.3. CREATE Command
create: C R E A T E SP mailbox;

// 6.3.4. DELETE Command
del: D E L E T E SP mailbox;

// 6.3.5. RENAME Command
rename: R E N A M E SP mailbox SP mailbox;

// 6.3.6. SUBSCRIBE Command
sub: S U B S C R I B E SP mailbox;

// 6.3.7. UNSUBSCRIBE Command
unsub: U N S U B S C R I B E SP mailbox;

// 6.3.8. LIST Command
list: L I S T SP mailbox SP listMailbox;

// 6.3.9. LSUB Command
lsub: L S U B SP mailbox SP listMailbox;

// 6.3.10. STATUS Command
status: S T A T U S SP mailbox SP LParen statusAtt (SP statusAtt)* RParen;

// RFC2177 IDLE Command
idle: I D L E;
done: D O N E;

statusAtt
	: M E S S A G E S
	| R E C E N T
	| U I D N E X T
	| U I D V A L I D I T Y
	| U N S E E N
	;

// 6.3.11. APPEND Command
append: A P P E N D SP mailbox (SP flagList)? (SP dateTime)? SP literal;

flagList: LParen (flag (SP flag)*)? RParen;

flag
	: Backslash A N S W E R E D
	| Backslash F L A G G E D
	| Backslash D E L E T E D
	| Backslash S E E N
	| Backslash D R A F T
	| flagKeyword 
	| flagExtension
	;

flagKeyword: atom;

flagExtension: Backslash atom;

dateTime: DQuote dateDayFixed Minus dateMonth Minus dateYear SP time SP zone DQuote;

dateDayFixed
	: SP digit
	| digit digit
	;

dateMonth: J A N | F E B | M A R | A P R | M A Y | J U N | J U L | A U G | S E P | O C T | N O V | D E C;

dateYear: digit digit digit digit;

time: digit digit Colon digit digit Colon digit digit;

zone: sign digit digit digit digit;

sign
	: Plus		# signPlus
	| Minus		# signMinus
	;

// 6.4. Client Commands - Selected State
commandSelected: check | close | expunge | uidExpunge | unselect | copy | move | fetch | store | uid | search;

// 6.4.1. CHECK Command
check: C H E C K;

// 6.4.2. CLOSE Command
close: C L O S E;

// 6.4.3. EXPUNGE Command
expunge: E X P U N G E;

// RFC4315 UIDPLUS Extension
uidExpunge: U I D SP E X P U N G E SP seqSet;

// RFC3691 UNSELECT Extension
unselect: U N S E L E C T;

// 6.4.4. SEARCH Command
search: S E A R C H (SP C H A R S E T SP astring)? (SP searchKey)+;

searchKey
	: A L L                                     # searchKeyAll
	| A N S W E R E D                           # searchKeyAnswered
	| B C C SP astring                          # searchKeyBcc
	| B E F O R E SP date                       # searchKeyBefore
	| B O D Y SP astring                        # searchKeyBody
	| C C SP astring                            # searchKeyCc
	| D E L E T E D                             # searchKeyDeleted
	| F L A G G E D                             # searchKeyFlagged
	| F R O M SP astring                        # searchKeyFrom
	| K E Y W O R D SP flagKeyword              # searchKeyKeyword
	| N E W                                     # searchKeyNew
	| O L D                                     # searchKeyOld
	| O N SP date                               # searchKeyOn
	| R E C E N T                               # searchKeyRecent
	| S E E N                                   # searchKeySeen
	| S I N C E SP date                         # searchKeySince
	| S U B J E C T SP astring                  # searchKeySubject
	| T E X T SP astring                        # searchKeyText
	| T O SP astring                            # searchKeyTo
	| U N A N S W E R E D                       # searchKeyUnanswered
	| U N D E L E T E D                         # searchKeyUndeleted
	| U N F L A G G E D                         # searchKeyUnflagged
	| U N K E Y W O R D SP flagKeyword          # searchKeyUnkeyword
	| U N S E E N                               # searchKeyUnseen
	| D R A F T                                 # searchKeyDraft
	| H E A D E R SP headerFieldName SP astring # searchKeyHeader
	| L A R G E R SP number                     # searchKeyLarger
	| N O T SP searchKey                        # searchKeyNot
	| O R SP searchKey SP searchKey             # searchKeyOr
	| S E N T B E F O R E SP date               # searchKeySentBefore
	| S E N T O N SP date                       # searchKeySentOn
	| S E N T S I N C E SP date                 # searchKeySentSince
	| S M A L L E R SP number                   # searchKeySmaller
	| U I D SP seqSet                           # searchKeyUID
	| U N D R A F T                             # searchKeyUndraft
	| seqSet                                    # searchKeySeqSet
	| LParen searchKey (SP searchKey)* RParen   # searchKeyList
	;

date
	: dateText
	| DQuote dateText DQuote
	;

dateText: dateDay Minus dateMonth Minus dateYear;

dateDay: digit | digit digit;

// 6.4.5. FETCH Command
fetch: F E T C H SP seqSet SP fetchTarget;

seqSet: seqItem (Comma seqItem)*;

seqItem
	: seqNumber 	# seqItemNum
	| seqRange		# seqItemRng
	;

seqNumber: number | Asterisk;

seqRange: seqNumber Colon seqNumber;

fetchTarget
	: A L L                                 # fetchTargetAll
	| F U L L                               # fetchTargetFull
	| F A S T                               # fetchTargetFast
	| fetchAtt                              # fetchTargetAtt
	| LParen fetchAtt (SP fetchAtt)* RParen # fetchTargetAtt
	;

fetchAtt
	: E N V E L O P E                                   # fetchAttEnvelope
	| F L A G S                                         # fetchAttFlags
	| I N T E R N A L D A T E                           # fetchAttInternalDate
	| R F C N8 N2 N2                                    # fetchAttRFC822
	| R F C N8 N2 N2 Period H E A D E R                 # fetchAttRFC822Header
	| R F C N8 N2 N2 Period S I Z E                     # fetchAttRFC822Size
	| R F C N8 N2 N2 Period T E X T                     # fetchAttRFC822Text
	| B O D Y                                           # fetchAttBody
	| B O D Y S T R U C T U R E                         # fetchAttBodyStructure
	| U I D                                             # fetchAttUID
	| B O D Y peek? LBracket section? RBracket partial? # fetchAttBodySection
	;

peek: Period P E E K;

partial: Less number Period number Greater;

section
	: sectionMsgText                      # sectionKeyword
	| sectionPart (Period sectionText)?   # sectionWithPart
	;

sectionMsgText
	: H E A D E R                                                   # sectionKwdHeader
	| T E X T                                                       # sectionKwdText
	| H E A D E R Period F I E L D S headerFieldsNot? SP headerList # sectionKwdHeaderFields
	;

headerFieldsNot: Period N O T;

headerList: LParen headerFieldName (SP headerFieldName)* RParen;

headerFieldName: astring;

sectionPart: number (Period number)*;

sectionText
  : sectionMsgText 
  | M I M E
  ;

// 6.4.6. STORE Command
store: S T O R E SP seqSet SP storeAction SP storeFlags;

storeAction: sign? F L A G S silent?;

silent: Period S I L E N T;

storeFlags
	: flagList 					# storeFlagList
	| flag (SP flag)*		# storeSpacedFlags
	;

// 6.4.7. COPY Command
copy: C O P Y SP seqSet SP mailbox;

// RFC6851 3.1 MOVE Command
move: M O V E SP seqSet SP mailbox;

// 6.4.8. UID Command
uid: U I D SP (copy | fetch | search | store | move);

// RFC 2971: IMAP ID extension
id: I D SP id_param_list;

id_param_list
    : id_nil_param
    | id_params
    ;

id_nil_param: N I L;

                // Restrict parsing rule to 30 fields
id_params : LParen (c+=id_param_key_pair SP?)* {$c.size() <=30}? RParen ;

id_param_key_pair: string SP id_param_key_value;

id_param_key_value
    : id_nil_param
    | nstring
    ;

// Common
digit: N0 | N1 | N2 | N3 | N4 | N5 | N6 | N7 | N8 | N9;

alpha: A | B | C | D | E | F | G | H | I | J | K | L | M | N | O | P | Q | R | S | T | U | V | W | X | Y | Z;

baseChar
	: alpha
	| digit
	| Ampersand
	| At
	| Backtick
	| Caret
	| Colon
	| Comma
	| Dollar
	| Equal
	| Exclamation
	| Greater
	| Hash
	| LBracket
	| Less
	| Minus
	| Period
	| Pipe
	| Question
	| RCurly
	| Semicolon
	| Slash
	| SQuote
	| Tilde
	| Underscore
	;

tagChar
	: baseChar
	| RBracket
	;

atomChar
	: baseChar
	| Plus
	;

unescapedChar
	: baseChar
	| U_01 | U_02 | U_03 | U_04 | U_05 | U_06 | U_07 | U_08 | U_09 | U_0B | U_0C | U_0E | U_0F
	| U_10 | U_11 | U_12 | U_13 | U_14 | U_15 | U_16 | U_17 | U_18 | U_19 | U_1A | U_1B | U_1C | U_1D | U_1E | U_1F
	| DEL
	| SP
	| Percent
	| LParen
	| RParen
	| Asterisk
	| Plus
	| RBracket
	| LCurly
	;

astringChar: atomChar | respSpecials;

respSpecials: RBracket;


parser grammar RFC5322AddressListParser;

options { tokenVocab=RFC5322AddressListLexer; }


quotedString: QuotedText;

word
    : CFWS? atom CFWS?
	| CFWS? quotedString CFWS?
    | encodedAtom
	;

// Allow dotAtom to have a trailing period; some messages in the wild look like this.
dotatom: Atom (Period Atom)* Period?;

atom: Atom;

encodedAtom: (CFWS? EncodedAtom)+ CFWS?;

// --------------------------
// 3.4. Address Specification
// --------------------------

address
	: mailbox
	| group
	;

mailbox
	: nameAddr
	| addrSpec
	;

nameAddr: displayName? angleAddr;

angleAddr
	: CFWS? Less addrSpec? Greater CFWS?
	| obsAngleAddr
	;

// relax this definition to allow the final semicolon to be optional
// and to permit it to be surrounded by quotes.
group: displayName Colon groupList? Semicolon? CFWS?;

unspaced: Period | At;

displayName
	: word+
	| word (word | unspaced | CFWS)*
	;

mailboxList
	: mailbox (Comma mailbox)*
	| obsMboxList
	;

addressList
	: address ((Comma | Semicolon) address)* EOF
	| obsAddrList EOF
	;

groupList
	: mailboxList
	| CFWS
	| obsGroupList
	;

// Allow addrSpec contain a port.
addrSpec: localPart At domain (Colon port)?;

port: Atom;

localPart
	: obsLocalPart
	| CFWS? dotatom CFWS?
	| CFWS? quotedString CFWS?
	;

domain
	: CFWS? dotatom CFWS?
	| CFWS? domainLiteral CFWS?
	| CFWS? obsDomain CFWS?
	;

domainLiteral: DomainLiteral;



// ------------------------
// 4.4. Obsolete Addressing
// ------------------------

obsAngleAddr: CFWS? Less obsRoute addrSpec Greater CFWS?;

obsRoute: obsDomainList Colon;

obsDomainList: (CFWS | Comma)* At domain (Comma CFWS? (At domain)?)*;

obsMboxList: (CFWS? Comma)* mailbox (Comma (mailbox | CFWS)?)*;

obsAddrList: (CFWS? Comma)* address (Comma (address | CFWS)?)*;

obsGroupList: (CFWS? Comma)+ CFWS?;

obsLocalPart: word (Period word)*;

obsDomain: atom (Period atom)*;





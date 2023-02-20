lexer grammar RFC5322AddressListLexer;

fragment U_00:             '\u0000';
fragment U_01_08:          '\u0001'..'\u0008';
fragment LF:               '\n';     // \u000A
fragment U_0B:             '\u000B';
fragment U_0C:             '\u000C';
fragment CR:               '\r';     // \u000D
fragment U_0E_1F:          '\u000E'..'\u001F';

// Printable (0x20-0x7E)
fragment Exclamation:      '!';      // \u0021
fragment DQuote:           '"';      // \u0022
fragment Hash:             '#';      // \u0023
fragment Dollar:           '$';      // \u0024
fragment Percent:          '%';      // \u0025
fragment Ampersand:        '&';      // \u0026
fragment SQuote:           '\'';     // \u0027
fragment LParens:          '(';      // \u0028
fragment RParens:          ')';      // \u0029
fragment Asterisk:         '*';      // \u002A
fragment Plus:             '+';      // \u002B
Comma:            ',';      // \u002C
fragment Minus:            '-';      // \u002D
Period:           '.';      // \u002E
fragment Slash:            '/';      // \u002F
fragment Digit:            [0-9];    // \u0030 -- \u0039
Colon:            ':';      // \u003A
Semicolon:        ';';      // \u003B
Less:             '<';      // \u003C
fragment Equal:            '=';      // \u003D
Greater:          '>';      // \u003E
fragment Question:         '?';      // \u003F
At:               '@';      // \u0040
// alphaUpper
fragment LBracket:         '[';      // \u005B
fragment Backslash:        '\\';     // \u005C
fragment RBracket:         ']';      // \u005D
fragment Caret:            '^';      // \u005E
fragment Underscore:       '_';      // \u005F
fragment Backtick:         '`';      // \u0060
// alphaLower
fragment LCurly:           '{';      // \u007B
fragment Pipe:             '|';      // \u007C
fragment RCurly:           '}';      // \u007D
fragment Tilde:            '~';      // \u007E

// Other
Delete: '\u007F';

// RFC6532 Extension
UTF8NonAscii: '\u0080'..'\uFFFF';

fragment A: 'A'|'a';
fragment B: 'B'|'b';
fragment C: 'C'|'c';
fragment D: 'D'|'d';
fragment E: 'E'|'e';
fragment F: 'F'|'f';
fragment G: 'G'|'g';
fragment H: 'H'|'h';
fragment I: 'I'|'i';
fragment J: 'J'|'j';
fragment K: 'K'|'k';
fragment L: 'L'|'l';
fragment M: 'M'|'m';
fragment N: 'N'|'n';
fragment O: 'O'|'o';
fragment P: 'P'|'p';
fragment Q: 'Q'|'q';
fragment R: 'R'|'r';
fragment S: 'S'|'s';
fragment T: 'T'|'t';
fragment U: 'U'|'u';
fragment V: 'V'|'v';
fragment W: 'W'|'w';
fragment X: 'X'|'x';
fragment Y: 'Y'|'y';
fragment Z: 'Z'|'z';

// -------------------
// 3.2. Lexical tokens
// -------------------

fragment QuotedChar: VChar | WSP;

fragment QuotedPair
	: Backslash QuotedChar
	| ObsQP
	;

CFWS
	: (FWS? Comment)+ FWS?
	| FWS
	;

FWS
	: (WSP* CRLF)? WSP+
	| ObsFWS
	;


fragment AText
	: Alpha
	| Digit
	| Exclamation
	| Hash
	| Dollar
	| Percent
	| Ampersand
	| SQuote
	| Asterisk
	| Plus
	| Minus
	| Slash
	| Equal
	| Question
	| Caret
	| Underscore
	| Backtick
	| LCurly
	| Pipe
	| RCurly
	| Tilde
	| UTF8NonAscii
	;

fragment QText
    : AText
	| LParens
	| RParens
	| Comma
	| Period
	| Digit
	| Colon
	| Semicolon
	| Less
	| Greater
	| At
	| LBracket
	| RBracket
	| ObsQtext
	;

fragment QuotedContent
	: QText
	| QuotedPair
	;

QuotedText: DQuote (FWS? QuotedContent)* FWS? DQuote;

EncodedAtom: Equal Question AText+ Question Equal;

Atom: AText+;


// --------------------------
// 3.4. Address Specification
// --------------------------

fragment DText
	: Alpha
	| Exclamation
	| DQuote
	| Hash
	| Dollar
	| Percent
	| Ampersand
	| SQuote
	| LParens
	| RParens
	| Asterisk
	| Plus
	| Comma
	| Minus
	| Period
	| Slash
	| Digit
	| Colon
	| Semicolon
	| Less
	| Equal
	| Greater
	| Question
	| At
	| Caret
	| Underscore
	| Backtick
	| LCurly
	| Pipe
	| RCurly
	| Tilde
//| obsDtext
	| UTF8NonAscii
	;

DomainLiteral: LBracket (FWS? DText)* FWS? RBracket;

// ----------------------------------
// 4.1. Miscellaneous Obsolete Tokens
// ----------------------------------

ObsNoWSCTL
	: U_01_08
	| U_0B
	| U_0C
	| U_0E_1F
	| Delete
	;

ObsCtext: ObsNoWSCTL;

ObsQtext: ObsNoWSCTL;

ObsQP: Backslash (U_00 | ObsNoWSCTL | LF | CR);

// ---------------------------------
// 4.2. Obsolete Folding White Space
// ---------------------------------

ObsFWS: WSP+ (CRLF WSP+);

// ---------------------------
// 4.3. Obsolete Date and Time
// ---------------------------

ObsZone
	: U T
	| U T C
	| G M T
	| E S T
	| E D T
	| C S T
	| C D T
	| M S T
	| M D T
	| P S T
	| P D T
//| obsZoneMilitary
	;

// -------------------------
// B.1. Core Rules (RFC5234)
// -------------------------

CRLF: CR LF;

WSP: ' '  | '\t' ;

fragment VChar
	: Alpha
	| Exclamation
	| DQuote
	| Hash
	| Dollar
	| Percent
	| Ampersand
	| SQuote
	| LParens
	| RParens
	| Asterisk
	| Plus
	| Comma
	| Minus
	| Period
	| Slash
	| Digit
	| Colon
	| Semicolon
	| Less
	| Equal
	| Greater
	| Question
	| At
	| LBracket
	| Backslash
	| RBracket
	| Caret
	| Underscore
	| Backtick
	| LCurly
	| Pipe
	| RCurly
	| Tilde
	| UTF8NonAscii
	;


fragment Alpha: A | B | C | D | E | F | G | H | I | J | K | L | M | N | O | P | Q | R | S | T | U | V | W | X | Y | Z ;

fragment CText
	: Alpha
	| Exclamation
	| DQuote
	| Hash
	| Dollar
	| Percent
	| Ampersand
	| SQuote
	| Asterisk
	| Plus
	| Comma
	| Minus
	| Period
	| Slash
	| Digit
	| Colon
	| Semicolon
	| Less
	| Equal
	| Greater
	| Question
	| At
	| LBracket
	| RBracket
	| Caret
	| Underscore
	| Backtick
	| LCurly
	| Pipe
	| RCurly
	| Tilde
	| ObsCtext
	| UTF8NonAscii
	;

CContent
	: CText
	| QuotedPair
	| Comment
	;

Comment: LParens (FWS? CContent)* FWS? RParens;

// ------------------------------------
// 2. Syntax of encoded-words (RFC2047)
// ------------------------------------


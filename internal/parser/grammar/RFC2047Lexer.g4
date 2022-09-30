lexer grammar RFC2047Lexer;

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
fragment Comma:            ',';      // \u002C
fragment Minus:            '-';      // \u002D
fragment Period:           '.';      // \u002E
fragment Slash:            '/';      // \u002F
fragment Digit:            [0-9];    // \u0030 -- \u0039
fragment Colon:            ':';      // \u003A
fragment Semicolon:        ';';      // \u003B
fragment Less:             '<';      // \u003C
fragment Equal:            '=';      // \u003D
fragment Greater:          '>';      // \u003E
Question:         '?';      // \u003F
fragment At:               '@';      // \u0040
fragment LBracket:         '[';      // \u005B
fragment Backslash:        '\\';     // \u005C
fragment RBracket:         ']';      // \u005D
fragment Caret:            '^';      // \u005E
fragment Underscore:       '_';      // \u005F
fragment Backtick:         '`';      // \u0060
fragment LCurly:           '{';      // \u007B
fragment Pipe:             '|';      // \u007C
fragment RCurly:           '}';      // \u007D
fragment Tilde:            '~';      // \u007E

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

fragment Alpha: A | B | C | D | E | F | G | H | I | J | K | L | M | N | O | P | Q | R | S | T | U | V | W | X | Y | Z ;

EncodeBegin: Equal Question;
EncodeEnd : Question Equal;

Encoding: Q|B;

fragment TokenChar
	: Alpha
	| Exclamation
	| Hash
	| Dollar
	| Percent
	| Ampersand
	| SQuote
	| Asterisk
	| Plus
	| Minus
	| Digit
	| Backslash
	| Caret
	| Underscore
	| Backtick
	| LCurly
	| Pipe
	| RCurly
	| Tilde
	;


Token: TokenChar+;

fragment EncodedChar
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
	;

EncodedText: EncodedChar+;

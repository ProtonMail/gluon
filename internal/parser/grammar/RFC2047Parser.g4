parser grammar RFC2047Parser;

options { tokenVocab=RFC2047Lexer; }

// ------------------------------------
// 2. Syntax of encoded-words (RFC2047)
// ------------------------------------

encodedWordList: encodedWord+;

encodedWord: Equal Question token Question encoding Question encodedText Question Equal;

encoding: Q | B;

token: tokenChar+;

alpha: A | B | C | D | E | F | G | H | I | J | K | L | M | N | O | P | Q | R | S | T | U | V | W | X | Y | Z ;

tokenChar
	: alpha
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

encodedChar
	: alpha
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

encodedText: encodedChar+;

parser grammar RFC2047Parser;

options { tokenVocab=RFC2047Lexer; }

// ------------------------------------
// 2. Syntax of encoded-words (RFC2047)
// ------------------------------------

encodedWordList: encodedWord+;

encodedWord: EncodeBegin Token Question Encoding Question encodedText EncodeEnd;

encodedText: EncodedText | Token;

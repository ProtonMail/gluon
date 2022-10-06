parser grammar RFC5322DateTimeParser;

options { tokenVocab=RFC5322DateTimeLexer; }


dateTime: (dayOfweek Comma)? day month year hour Colon minute (Colon second)?  FWS? zone? cfws? EOF;

dayOfweek
	: FWS? dayName
	| cfws? dayName cfws?
	;

dayName: Day;

day
	: FWS? Digit Digit? FWS
	| cfws? Digit Digit? cfws?
	;

month: Month;

year
	: cfws? Digit Digit cfws?
	| FWS Digit Digit Digit Digit FWS
	;

// NOTE: RFC5322 requires two digits for the hour, but we
// relax that requirement a bit, allowing single digits.
hour
	: Digit? Digit
	| cfws? Digit? Digit cfws?
	;

minute
	: Digit Digit
	| cfws? Digit Digit cfws?
	;

second
	: Digit Digit
	| cfws? Digit Digit cfws?
	;

offset: (Plus | Minus)? Digit Digit Digit Digit;

zone
	: ObsZone
	| offset
	;

cfws
	: (FWS? Comment)+ FWS?
	| FWS
	;

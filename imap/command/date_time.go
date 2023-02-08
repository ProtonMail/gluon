package command

import (
	"fmt"
	"github.com/ProtonMail/gluon/imap/parser"
	"time"
)

func ParseDateTime(p *parser.Parser) (time.Time, error) {
	//  date-time       = DQUOTE date-day-fixed "-" date-month "-" date-year SP time SP zone DQUOTE
	if err := p.Consume(parser.TokenTypeDQuote, `Expected '"' at start of date time`); err != nil {
		return time.Time{}, err
	}

	dateDay, err := ParseDateDayFixed(p)
	if err != nil {
		return time.Time{}, err
	}

	if err := p.Consume(parser.TokenTypeMinus, `Expected '-' after date day`); err != nil {
		return time.Time{}, err
	}

	dateMonth, err := ParseDateMonth(p)
	if err != nil {
		return time.Time{}, err
	}

	if err := p.Consume(parser.TokenTypeMinus, `Expected '-' after date month`); err != nil {
		return time.Time{}, err
	}

	dateYear, err := ParseDateYear(p)
	if err != nil {
		return time.Time{}, err
	}

	if err := p.Consume(parser.TokenTypeSP, `Expected space after date year`); err != nil {
		return time.Time{}, err
	}

	timeHour, timeMin, timeSec, err := ParseTime(p)
	if err != nil {
		return time.Time{}, err
	}

	if err := p.Consume(parser.TokenTypeSP, `Expected space after date time`); err != nil {
		return time.Time{}, err
	}

	timeZone, err := ParseZone(p)
	if err != nil {
		return time.Time{}, err
	}

	if err := p.Consume(parser.TokenTypeDQuote, `Expected '"' at end of date time`); err != nil {
		return time.Time{}, err
	}

	return time.Date(dateYear, dateMonth, dateDay, timeHour, timeMin, timeSec, 0, timeZone), nil
}

func ParseDateDayFixed(p *parser.Parser) (int, error) {
	// date-day-fixed  = (SP DIGIT) / 2DIGIT
	if ok, err := p.Matches(parser.TokenTypeSP); err != nil {
		return 0, err
	} else if ok {
		if err := p.Consume(parser.TokenTypeDigit, "expected digit after space separated date day"); err != nil {
			return 0, err
		}

		return parser.ByteToInt(p.PreviousToken().Value), nil
	}

	return p.ParseNumberN(2)
}

var dateMonthToInt = map[string]time.Month{
	"Jan": time.January,
	"Feb": time.February,
	"Mar": time.March,
	"Apr": time.April,
	"May": time.May,
	"Jun": time.June,
	"Jul": time.July,
	"Aug": time.August,
	"Sep": time.September,
	"Oct": time.October,
	"Nov": time.November,
	"Dec": time.December,
}

func ParseDateMonth(p *parser.Parser) (time.Month, error) {
	month := make([]byte, 3)

	for i := 0; i < 3; i++ {
		if err := p.Consume(parser.TokenTypeChar, "unexpected character for date month"); err != nil {
			return 0, err
		}

		month[i] = p.PreviousToken().Value
	}

	v, ok := dateMonthToInt[string(month)]
	if !ok {
		return 0, p.MakeError(fmt.Sprintf("invalid date month '%v'", string(month)))
	}

	return v, nil
}

func ParseDateYear(p *parser.Parser) (int, error) {
	return p.ParseNumberN(4)
}

func ParseZone(p *parser.Parser) (*time.Location, error) {
	multiplier := 1

	if ok, err := p.Matches(parser.TokenTypePlus); err != nil {
		return nil, err
	} else if !ok {
		if ok, err := p.Matches(parser.TokenTypeMinus); err != nil {
			return nil, err
		} else if ok {
			multiplier = -1
		} else {
			return nil, p.MakeError("expected either '+' or '-' on time zone start")
		}
	}

	zoneHour, err := p.ParseNumberN(2)
	if err != nil {
		return nil, err
	}

	zoneMinute, err := p.ParseNumberN(2)
	if err != nil {
		return nil, err
	}

	zone := (zoneHour*3600 + zoneMinute*60) * multiplier

	return time.FixedZone("zone", zone), nil
}

func ParseTime(p *parser.Parser) (int, int, int, error) {
	hour, err := p.ParseNumberN(2)
	if err != nil {
		return 0, 0, 0, err
	}

	if err := p.Consume(parser.TokenTypeColon, "expected colon after hour component"); err != nil {
		return 0, 0, 0, err
	}

	min, err := p.ParseNumberN(2)
	if err != nil {
		return 0, 0, 0, err
	}

	if err := p.Consume(parser.TokenTypeColon, "expected colon after minute component"); err != nil {
		return 0, 0, 0, err
	}

	sec, err := p.ParseNumberN(2)
	if err != nil {
		return 0, 0, 0, err
	}

	return hour, min, sec, nil
}

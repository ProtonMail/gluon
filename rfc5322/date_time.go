package rfc5322

import (
	"fmt"
	"time"

	"github.com/ProtonMail/gluon/rfcparser"
)

func parseDTDateTime(p *rfcparser.Parser) (time.Time, error) {
	//  date-time       =   [ day-of-week "," ] date time [CFWS]
	if _, err := tryParseCFWS(p); err != nil {
		return time.Time{}, err
	}

	if p.Check(rfcparser.TokenTypeChar) {
		if err := parseDTDayOfWeek(p); err != nil {
			return time.Time{}, err
		}

		if err := p.Consume(rfcparser.TokenTypeComma, "expected ',' after day of the week"); err != nil {
			return time.Time{}, err
		}
	}

	year, month, day, err := parseDTDate(p)
	if err != nil {
		return time.Time{}, err
	}

	hour, min, sec, zone, err := parseDTTime(p)
	if err != nil {
		return time.Time{}, err
	}

	if _, err := tryParseCFWS(p); err != nil {
		return time.Time{}, err
	}

	return time.Date(year, month, day, hour, min, sec, 0, zone), nil
}

func parseDTDayOfWeek(p *rfcparser.Parser) error {
	// nolint:dupword
	// day-of-week     =   ([FWS] day-name) / obs-day-of-week
	// obs-day-of-week =   [CFWS] day-name [CFWS]
	//
	if _, err := tryParseCFWS(p); err != nil {
		return err
	}

	dayBytes, err := p.CollectBytesWhileMatches(rfcparser.TokenTypeChar)
	if err != nil {
		return err
	}

	dayStr := dayBytes.IntoString().ToLower()

	_, ok := dateDaySet[dayStr.Value]
	if !ok {
		return p.MakeErrorAtOffset(fmt.Sprintf("invalid day name '%v'", dayStr.Value), dayBytes.Offset)
	}

	if _, err := tryParseCFWS(p); err != nil {
		return err
	}

	return nil
}

// Return (year, month, day).
func parseDTDate(p *rfcparser.Parser) (int, time.Month, int, error) {
	day, err := parseDTDay(p)
	if err != nil {
		return 0, 0, 0, err
	}

	month, err := parseDTMonth(p)
	if err != nil {
		return 0, 0, 0, err
	}

	year, err := parseDTYear(p)
	if err != nil {
		return 0, 0, 0, err
	}

	return year, month, day, nil
}

func parseDTDay(p *rfcparser.Parser) (int, error) {
	// day             =   ([FWS] 1*2DIGIT FWS) / obs-day
	//
	// obs-day         =   [CFWS] 1*2DIGIT [CFWS]
	//
	if _, err := tryParseCFWS(p); err != nil {
		return 0, err
	}

	if err := p.Consume(rfcparser.TokenTypeDigit, "expected digit for day value"); err != nil {
		return 0, err
	}

	day := rfcparser.ByteToInt(p.PreviousToken().Value)

	if ok, err := p.Matches(rfcparser.TokenTypeDigit); err != nil {
		return 0, err
	} else if ok {
		day *= 10
		day += rfcparser.ByteToInt(p.PreviousToken().Value)
	}

	if _, err := tryParseCFWS(p); err != nil {
		return 0, err
	}

	return day, nil
}

func parseDTMonth(p *rfcparser.Parser) (time.Month, error) {
	// month           =   "Jan" / "Feb" / "Mar" / "Apr" /
	//                     "May" / "Jun" / "Jul" / "Aug" /
	//                     "Sep" / "Oct" / "Nov" / "Dec"
	//
	month := make([]byte, 3)

	for i := 0; i < 3; i++ {
		if err := p.Consume(rfcparser.TokenTypeChar, "unexpected character for date month"); err != nil {
			return 0, err
		}

		month[i] = p.PreviousToken().Value
	}

	v, ok := dateMonthToTimeMonth[string(month)]
	if !ok {
		return 0, p.MakeError(fmt.Sprintf("invalid date month '%v'", string(month)))
	}

	return v, nil
}

func parseDTYear(p *rfcparser.Parser) (int, error) {
	// year            =   (FWS 4*DIGIT FWS) / obs-year
	//
	// obs-year        =   [CFWS] 2*DIGIT [CFWS]
	//
	if _, err := tryParseCFWS(p); err != nil {
		return 0, err
	}

	year, err := p.ParseNumberN(2)
	if err != nil {
		return 0, err
	}

	if p.Check(rfcparser.TokenTypeDigit) {
		yearPart2, err := p.ParseNumberN(2)
		if err != nil {
			return 0, err
		}

		year *= 100
		year += yearPart2
	} else {
		if year > time.Now().Year()%100 {
			year += 1900
		} else {
			year += 2000
		}
	}

	if _, err := tryParseCFWS(p); err != nil {
		return 0, err
	}

	return year, nil
}

func parseDTTime(p *rfcparser.Parser) (int, int, int, *time.Location, error) {
	// time            =   time-of-day zone
	//
	hour, min, sec, err := parseDTTimeOfDay(p)
	if err != nil {
		return 0, 0, 0, nil, err
	}

	loc, err := parseDTZone(p)
	if err != nil {
		return 0, 0, 0, nil, err
	}

	return hour, min, sec, loc, nil
}

func parseDTTimeOfDay(p *rfcparser.Parser) (int, int, int, error) {
	// time-of-day     =   hour ":" minute [ ":" second ]
	hour, err := parseDTHour(p)
	if err != nil {
		return 0, 0, 0, err
	}

	if err := p.Consume(rfcparser.TokenTypeColon, "expected ':' after hour"); err != nil {
		return 0, 0, 0, err
	}

	min, err := parseDTMin(p)
	if err != nil {
		return 0, 0, 0, err
	}

	var sec int

	if ok, err := p.Matches(rfcparser.TokenTypeColon); err != nil {
		return 0, 0, 0, err
	} else if ok {
		s, err := parseDTSecond(p)
		if err != nil {
			return 0, 0, 0, err
		}

		sec = s
	}

	return hour, min, sec, nil
}

func parseDTHour(p *rfcparser.Parser) (int, error) {
	return parseDT2Digit(p)
}

func parseDTMin(p *rfcparser.Parser) (int, error) {
	return parseDT2Digit(p)
}

func parseDTSecond(p *rfcparser.Parser) (int, error) {
	return parseDT2Digit(p)
}

func parseDT2Digit(p *rfcparser.Parser) (int, error) {
	// 2digit          =   2DIGIT / obs-second
	//
	// obs-2digit      =   [CFWS] 2DIGIT [CFWS]
	//
	if _, err := tryParseCFWS(p); err != nil {
		return 0, err
	}

	num, err := p.ParseNumberN(2)
	if err != nil {
		return 0, err
	}

	if _, err := tryParseCFWS(p); err != nil {
		return 0, err
	}

	return num, nil
}

func parseDTZone(p *rfcparser.Parser) (*time.Location, error) {
	// zone            =   (FWS ( "+" / "-" ) 4DIGIT) / obs-zone
	//
	//     obs-zone        =   "UT" / "GMT" /     ; Universal Time
	//                                            ; North American UT
	//                                            ; offsets
	//                         "EST" / "EDT" /    ; Eastern:  - 5/ - 4
	//                         "CST" / "CDT" /    ; Central:  - 6/ - 5
	//                         "MST" / "MDT" /    ; Mountain: - 7/ - 6
	//                         "PST" / "PDT" /    ; Pacific:  - 8/ - 7
	//                                            ;
	if _, err := tryParseCFWS(p); err != nil {
		return nil, err
	}

	if !(p.Check(rfcparser.TokenTypeDigit) || p.Check(rfcparser.TokenTypeMinus) || p.Check(rfcparser.TokenTypePlus) || p.Check(rfcparser.TokenTypeChar)) {
		return time.UTC, nil
	}

	multiplier := 1

	if ok, err := p.Matches(rfcparser.TokenTypePlus); err != nil {
		return nil, err
	} else if !ok {
		if ok, err := p.Matches(rfcparser.TokenTypeMinus); err != nil {
			return nil, err
		} else if ok {
			multiplier = -1
		} else if !(p.Check(rfcparser.TokenTypeDigit) || p.Check(rfcparser.TokenTypeChar)) {
			return nil, p.MakeError("expected either '+' or '-' on time zone start")
		}
	}

	// New format.
	if p.Check(rfcparser.TokenTypeDigit) {
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

	// Old Format

	value, err := p.CollectBytesWhileMatches(rfcparser.TokenTypeChar)
	if err != nil {
		return nil, err
	}

	valueStr := value.IntoString().ToLower()

	loc, ok := obsZoneToLocation[valueStr.Value]
	if !ok {
		return nil, p.MakeErrorAtOffset(fmt.Sprintf("unknown time zone '%v'", valueStr), value.Offset)
	}

	if _, err := tryParseCFWS(p); err != nil {
		return nil, err
	}

	return loc, nil
}

var obsZoneToLocation = map[string]*time.Location{
	"ut":  time.FixedZone("ut", 0),
	"gmt": time.FixedZone("gmt", 0),
	"utc": time.FixedZone("utc", 0),
	"est": time.FixedZone("est", -5*60*60),
	"edt": time.FixedZone("edt", -4*60*60),
	"cst": time.FixedZone("cst", -6*60*60),
	"cdt": time.FixedZone("cdt", -5*60*60),
	"mst": time.FixedZone("mst", -7*60*60),
	"mdt": time.FixedZone("mdt", -6*60*60),
	"pst": time.FixedZone("pst", -8*60*60),
	"pdt": time.FixedZone("pdt", -7*60*60),
}

var dateMonthToTimeMonth = map[string]time.Month{
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

var dateDaySet = map[string]struct{}{
	"mon": {},
	"tue": {},
	"wed": {},
	"thu": {},
	"fri": {},
	"sat": {},
	"sun": {},
}

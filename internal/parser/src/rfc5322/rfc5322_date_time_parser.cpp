#include "rfc5322_date_time_parser.h"

#include "RFC5322DateTimeLexer.h"
#include "RFC5322DateTimeParser.h"
#include "RFC5322DateTimeParserBaseVisitor.h"
#include "error_listener.h"

namespace rfc5322 {

static constexpr int char_to_digit(char c) {
  return int(c - '0');
}

static inline int combineDigits(std::vector<antlr4::tree::TerminalNode*> vector) {
  int result = 0;

  if (!vector.empty()) {
    result = char_to_digit(vector[0]->getText()[0]);
    auto itEnd = vector.end();
    for (auto it = vector.begin() + 1; it != itEnd; it++) {
      result *= 10;
      result += char_to_digit((*it)->getText()[0]);
    }
  }

  return result;
}

class DateTimeVisitor final : public RFC5322DateTimeParserBaseVisitor {
 public:
  std::any visitDayOfweek(RFC5322DateTimeParser::DayOfweekContext* ctx) override { return std::any{}; }

  std::any visitDayName(RFC5322DateTimeParser::DayNameContext* ctx) override { return std::any{}; }

  std::any visitDay(RFC5322DateTimeParser::DayContext* ctx) override {
    mDateTime.day = combineDigits(ctx->Digit());
    return std::any{};
  }

  std::any visitMonth(RFC5322DateTimeParser::MonthContext* ctx) override {
    auto month = ctx->getText();
    for (auto& ch : month) {
      ch = tolower(ch);
    }

    if (month == "jan") {
      mDateTime.month = 1;
    } else if (month == "feb") {
      mDateTime.month = 2;
    } else if (month == "mar") {
      mDateTime.month = 3;
    } else if (month == "apr") {
      mDateTime.month = 4;
    } else if (month == "may") {
      mDateTime.month = 5;
    } else if (month == "jun") {
      mDateTime.month = 6;
    } else if (month == "jul") {
      mDateTime.month = 7;
    } else if (month == "aug") {
      mDateTime.month = 8;
    } else if (month == "sep") {
      mDateTime.month = 9;
    } else if (month == "oct") {
      mDateTime.month = 10;
    } else if (month == "nov") {
      mDateTime.month = 11;
    } else if (month == "dec") {
      mDateTime.month = 12;
    } else {
      throw std::runtime_error("Invalid month");
    }
    return std::any{};
  }

  std::any visitYear(RFC5322DateTimeParser::YearContext* ctx) override {
    auto digits = ctx->Digit();
    auto digitsSize = digits.size();
    mDateTime.year = combineDigits(std::move(digits));

    if (digitsSize <= 2) {
      std::time_t t = std::time(nullptr);
      std::tm* const tinfo = std::localtime(&t);
      if (mDateTime.year > tinfo->tm_year % 100) {
        mDateTime.year += 1900;
      } else {
        mDateTime.year += 2000;
      }
    }

    return std::any{};
  }

  std::any visitHour(RFC5322DateTimeParser::HourContext* ctx) override {
    mDateTime.hour = combineDigits(ctx->Digit());
    return std::any{};
  }

  std::any visitMinute(RFC5322DateTimeParser::MinuteContext* ctx) override {
    mDateTime.min = combineDigits(ctx->Digit());
    return std::any{};
  }

  std::any visitSecond(RFC5322DateTimeParser::SecondContext* ctx) override {
    mDateTime.sec = combineDigits(ctx->Digit());
    return std::any{};
  }

  std::any visitOffset(RFC5322DateTimeParser::OffsetContext* ctx) override { return std::any{}; }

  std::any visitZone(RFC5322DateTimeParser::ZoneContext* ctx) override {
    if (auto offset = ctx->offset(); offset != nullptr) {
      int hourHigh = char_to_digit(offset->Digit(0)->getText()[0]);
      int hourLow = char_to_digit(offset->Digit(1)->getText()[0]);
      int minHigh = char_to_digit(offset->Digit(2)->getText()[0]);
      int minLow = char_to_digit(offset->Digit(3)->getText()[0]);

      int hour = hourHigh * 10 + hourLow;
      int min = minHigh * 10 + minLow;
      bool positive = offset->Minus() == nullptr;
      mDateTime.tzOffset = EncodeOffset(positive, hour, min);
      mDateTime.tzType = TimeZoneType::Offset;
    } else {
      auto zoneText = ctx->ObsZone()->getText();
      for (auto& ch : zoneText) {
        ch = tolower(ch);
      }

      mDateTime.tzType = TimeZoneType::Code;
      if (zoneText == "ut") {
        mDateTime.tz = TimeZone::UT;
      } else if (zoneText == "utc") {
        mDateTime.tz = TimeZone::UTC;
      } else if (zoneText == "gmt") {
        mDateTime.tz = TimeZone::GMT;
      } else if (zoneText == "est") {
        mDateTime.tz = TimeZone::EST;
      } else if (zoneText == "edt") {
        mDateTime.tz = TimeZone::EDT;
      } else if (zoneText == "cst") {
        mDateTime.tz = TimeZone::CST;
      } else if (zoneText == "cdt") {
        mDateTime.tz = TimeZone::CDT;
      } else if (zoneText == "mst") {
        mDateTime.tz = TimeZone::MST;
      } else if (zoneText == "mdt") {
        mDateTime.tz = TimeZone::MDT;
      } else if (zoneText == "pst") {
        mDateTime.tz = TimeZone::PST;
      } else if (zoneText == "pdt") {
        mDateTime.tz = TimeZone::PDT;
      } else {
        throw std::runtime_error("Invalid time zone");
      }
    }
    return std::any{};
  }

  DateTime getDateTime() const { return mDateTime; }

 private:
  DateTime mDateTime = DateTime{
      0, 0, 0, 0, 0, 0, TimeZone::UTC, TimeZoneType::Code,
  };
};

DateTime ParseDateTime(std::string_view str) {
  antlr4::ANTLRInputStream inputStream{str};
  RFC5322DateTimeLexer lexer{&inputStream};
  antlr4::CommonTokenStream tokens{&lexer};
  RFC5322DateTimeParser parser{&tokens};

  lexer.removeErrorListeners();
  parser.removeErrorListeners();

  parser::ErrorListener errorListener;
  parser.addErrorListener(&errorListener);

  auto dateTimeContext = parser.dateTime();

  if (errorListener.didError()) {
    throw std::runtime_error(errorListener.m_errorMessage.value());
  }

  auto visitor = DateTimeVisitor();
  visitor.visitDateTime(dateTimeContext);

  return visitor.getDateTime();
}

uint32_t EncodeOffset(bool plus, uint8_t hour, uint8_t min) {
  uint32_t result = 0;
  result |= hour << 8;
  result |= min;
  result |= plus ? 1 << 31 : 0;
  return result;
}

void DecodeOffset(uint32_t input, char& sign, uint8_t& hour, uint8_t& min) {
  if (input & (1 << 31)) {
    sign = '+';
  } else {
    sign = '-';
  }

  hour = (input >> 8) & 0xFF;
  min = (input & 0xFF);
}

}  // namespace rfc5322

#include <gtest/gtest.h>

#include <rfc5322/rfc5322_date_time_parser.h>
#include <rfc5322/rfc5322_parser_capi.h>

#include <memory>

using namespace rfc5322;

struct TestInput {
  const char* input;
  rfc5322::DateTime expected;
};

void validateDTTest(const TestInput& test) {
  SCOPED_TRACE(test.input);
  rfc5322::DateTime dt;
  EXPECT_NO_THROW(dt = rfc5322::ParseDateTime(test.input));
  EXPECT_EQ(test.expected.day, dt.day);
  EXPECT_EQ(test.expected.month, dt.month);
  EXPECT_EQ(test.expected.year, dt.year);
  EXPECT_EQ(test.expected.hour, dt.hour);
  EXPECT_EQ(test.expected.min, dt.min);
  EXPECT_EQ(test.expected.sec, dt.sec);
  EXPECT_EQ(test.expected.tzType, dt.tzType);
  EXPECT_EQ(test.expected.tz, dt.tz);
  EXPECT_EQ(test.expected.tzOffset, dt.tzOffset);
}

static inline rfc5322::DateTime
dtWithTzCode(int day, int month, int year, int hour, int min, int sec, rfc5322::TimeZone tz) {
  return {day, month, year, hour, min, sec, tz, rfc5322::TimeZoneType::Code, 0};
}

static inline rfc5322::DateTime
dtWithTzOffset(int day, int month, int year, int hour, int min, int sec, uint32_t tzOffset) {
  return {day, month, year, hour, min, sec, rfc5322::TimeZone::UTC, rfc5322::TimeZoneType::Offset, tzOffset};
}

TEST(DateTime, OffsetEncodeDecode) {
  uint8_t hour = 12;
  uint8_t min = 54;

  auto offsetPositive = rfc5322::EncodeOffset(true, hour, min);
  auto offsetNegative = rfc5322::EncodeOffset(false, hour, min);

  {
    char sign;
    uint8_t decodedHour, decodedMin;
    rfc5322::DecodeOffset(offsetPositive, sign, decodedHour, decodedMin);
    EXPECT_EQ(sign, '+');
    EXPECT_EQ(hour, decodedHour);
    EXPECT_EQ(min, decodedMin);
  }

  {
    char sign;
    uint8_t decodedHour, decodedMin;
    rfc5322::DecodeOffset(offsetNegative, sign, decodedHour, decodedMin);
    EXPECT_EQ(sign, '-');
    EXPECT_EQ(hour, decodedHour);
    EXPECT_EQ(min, decodedMin);
  }
}

TEST(DateTime, ValidDates) {
  const TestInput inputs[] = {
      {
          "Fri, 21 Nov 1997 09:55:06",
          dtWithTzCode(21, 11, 1997, 9, 55, 6, rfc5322::TimeZone::UTC),
      },
      {
          "Fri, 21 Nov 1997 09:55:06 -0600",
          dtWithTzOffset(21, 11, 1997, 9, 55, 6, EncodeOffset(false, 6, 0)),
      },
      {
          "Tue, 1 Jul 2003 10:52:37 +0200",
          dtWithTzOffset(1, 07, 2003, 10, 52, 37, EncodeOffset(true, 2, 0)),
      },
      {
          "Thu, 13 Feb 1969 23:32:54 -0330",
          dtWithTzOffset(13, 02, 1969, 23, 32, 54, EncodeOffset(false, 3, 30)),
      },
      {
          "Thu, 13 Feb 1969 23:32 -0330 (Newfoundland Time)",
          dtWithTzOffset(13, 02, 1969, 23, 32, 0, EncodeOffset(false, 3, 30)),
      },
      {
          "2 Jan 2006 15:04:05 -0700",
          dtWithTzOffset(2, 1, 2006, 15, 04, 05, EncodeOffset(false, 7, 0)),
      },
      {
          "2 Jan 2006 15:04:05 MST",
          dtWithTzCode(2, 1, 2006, 15, 04, 05, TimeZone::MST),
      },
      {
          "2 Jan 2006 15:04 -0700",
          dtWithTzOffset(2, 1, 2006, 15, 04, 00, EncodeOffset(false, 7, 0)),
      },
      {
          "2 Jan 2006 15:04 MST",
          dtWithTzCode(2, 1, 2006, 15, 04, 0, TimeZone::MST),
      },
      {
          "2 Jan 06 15:04:05 -0700",
          dtWithTzOffset(2, 1, 2006, 15, 04, 05, EncodeOffset(false, 7, 0)),
      },
      {
          "2 Jan 06 15:04:05 MST",
          dtWithTzCode(2, 1, 2006, 15, 04, 05, TimeZone::MST),
      },
      {
          "Mon, 2 Jan 2006 15:04:05 -0700",
          dtWithTzOffset(2, 1, 2006, 15, 04, 05, EncodeOffset(false, 7, 0)),
      },
  };

  for (const auto& test : inputs) {
    validateDTTest(test);
  }
}

TEST(DateTime, Obselete) {
  const TestInput inputs[] = {
      {
          "21 Nov 97 09:55:06 GMT",
          dtWithTzCode(21, 11, 1997, 9, 55, 6, rfc5322::TimeZone::GMT),
      },
      {
          "Wed, 01 Jan 2020 12:00:00 UTC",
          dtWithTzCode(1, 1, 2020, 12, 0, 0, rfc5322::TimeZone::UTC),
      },
      {
          "Wed, 01 Jan 2020 13:00:00 UTC",
          dtWithTzCode(1, 1, 2020, 13, 0, 0, rfc5322::TimeZone::UTC),
      },
      {
          "Wed, 01 Jan 2020 12:30:00 UTC",
          dtWithTzCode(1, 1, 2020, 12, 30, 0, rfc5322::TimeZone::UTC),
      },
  };

  for (const auto& test : inputs) {
    validateDTTest(test);
  }
}

TEST(DateTime, Relaxed) {
  const TestInput inputs[] = {
      {
          "Mon, 28 Jan 2019 20:59:01 0000",
          dtWithTzOffset(28, 1, 2019, 20, 59, 01, EncodeOffset(true, 0, 0)),
      },
      {
          "Mon, 25 Sep 2017 5:25:40 +0200",
          dtWithTzOffset(25, 9, 2017, 5, 25, 40, EncodeOffset(true, 2, 0)),
      },
  };

  for (const auto& test : inputs) {
    validateDTTest(test);
  }
}

TEST(DateTime, Rejected) {
  const char* input = "Mon, 25 Sep 2017 5:25:40 +02";
  EXPECT_ANY_THROW(rfc5322::ParseDateTime(input));
  RFC5322DateTime dt;
  EXPECT_EQ(-1, RFC5322DateTime_parse(&dt, input));
}

TEST(DateTime, CAPI) {
  const char* offset_input =  "2 Jan 06 15:04:05 -0700";
  const char* tz_input =  "2 Jan 06 15:04:05 EST";

  {
    RFC5322DateTime dt;
    EXPECT_EQ(0, RFC5322DateTime_parse(&dt, offset_input));
    EXPECT_EQ(2, dt.day);
    EXPECT_EQ(1, dt.month);
    EXPECT_EQ(2006, dt.year);
    EXPECT_EQ(15, dt.hour);
    EXPECT_EQ(04, dt.min);
    EXPECT_EQ(05, dt.sec);
    EXPECT_EQ(TZ_TYPE_OFFSET, dt.tzType);
    EXPECT_EQ(EncodeOffset(false, 7, 0), dt.tz);
  }
  {
    RFC5322DateTime dt;
    EXPECT_EQ(0, RFC5322DateTime_parse(&dt, tz_input));
    EXPECT_EQ(2, dt.day);
    EXPECT_EQ(1, dt.month);
    EXPECT_EQ(2006, dt.year);
    EXPECT_EQ(15, dt.hour);
    EXPECT_EQ(04, dt.min);
    EXPECT_EQ(05, dt.sec);
    EXPECT_EQ(TZ_TYPE_CODE, dt.tzType);
    EXPECT_EQ(TZ_CODE_EST, dt.tz);
  }
}



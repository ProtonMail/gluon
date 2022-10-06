#pragma once

#include <string_view>

namespace rfc5322 {

enum class TimeZone {
  UT,
  UTC,
  GMT,
  EST,
  EDT,
  CST,
  CDT,
  MST,
  MDT,
  PST,
  PDT
};

enum class TimeZoneType {
  Offset,
  Code,
};

struct DateTime {
  int day;
  int month;
  int year;
  int hour;
  int min;
  int sec;
  TimeZone tz;
  TimeZoneType tzType;
  uint32_t tzOffset;
};


DateTime ParseDateTime(std::string_view);

uint32_t EncodeOffset(bool plus, uint8_t hour, uint8_t min);

void DecodeOffset(uint32_t input, char& sign, uint8_t& hour, uint8_t& min);

}
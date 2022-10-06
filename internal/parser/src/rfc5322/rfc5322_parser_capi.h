#pragma once
#include <inttypes.h>

#ifdef __cplusplus
extern "C" {
#endif

struct RFC5322AddressList;
typedef struct RFC5322AddressList RFC5322AddressList;

struct RFC5322Address {
  const char* name;
  const char* address;
};

RFC5322AddressList* RFC5322AddressList_new();

void RFC5322AddressList_free(RFC5322AddressList* list);

// Parsers the input string and returns the number of addresses in the message. On error returns -1.
// Use `rfc5322AddressList_error_str()` to get the error string;
int RFC5322AddressList_parse(RFC5322AddressList* list, const char* input);

// Get the textual representation of the last error or NULL if no error.
const char* RFC5322AddressList_error_str(RFC5322AddressList* list);

// Get the nth address that was parsed after a call to `RFC5322AddressList_parse()`. Note that this value is only
// valid while the instance of RFC5322AddressList is still alive and no further calls to `rfc5332Parser_parse()` have
// been made.
struct RFC5322Address RFC5322AddressList_get(RFC5322AddressList* list, int index);

enum TzType {
  TZ_TYPE_OFFSET = 0,
  TZ_TYPE_CODE,
};

enum TzCode {
  TZ_CODE_UT,
  TZ_CODE_UTC,
  TZ_CODE_GMT,
  TZ_CODE_EST,
  TZ_CODE_EDT,
  TZ_CODE_CST,
  TZ_CODE_CDT,
  TZ_CODE_MST,
  TZ_CODE_MDT,
  TZ_CODE_PST,
  TZ_CODE_PDT
};

struct RFC5322DateTime {
  int day;
  int month;
  int year;
  int hour;
  int min;
  int sec;
  int tzType;
  uint32_t tz; // Offset encoded (1 << 32) & tz for positive sign, (tz >> 8) &0xF for hours and (tz &0xFF) for minutes
};

typedef struct RFC5322DateTime RFC5322DateTime;

// Parse Date Time String, returns 0 on success -1 on failure.
int RFC5322DateTime_parse(RFC5322DateTime* output, const char* input);

#ifdef __cplusplus
}
#endif

#include "rfc5322_parser_capi.h"
#include "rfc5322_address_list_parser.h"
#include "rfc5322_date_time_parser.h"

extern "C" {

struct RFC5322AddressList {
  std::vector<rfc5322::Address> results;
  std::string errorString;
};

RFC5322AddressList* RFC5322AddressList_new() {
  return new RFC5322AddressList{};
}

void RFC5322AddressList_free(RFC5322AddressList* list) {
  delete list;
}

int RFC5322AddressList_parse(RFC5322AddressList* list, const char* input) {
  list->errorString.clear();
  list->results.clear();

  try {
    list->results = rfc5322::ParseAddressList(input);
  } catch (const std::exception& e) {
    list->errorString = e.what();
    return -1;
  }

  return int(list->results.size());
}

const char* RFC5322AddressList_error_str(RFC5322AddressList* list) {
  return list->errorString.empty() ? nullptr : list->errorString.c_str();
}

RFC5322Address RFC5322AddressList_get(RFC5322AddressList* list, int index) {
  if (index < 0 || size_t(index) >= list->results.size()) {
    return RFC5322Address{nullptr, nullptr};
  }

  const auto& addr = list->results[index];
  return RFC5322Address{
      addr.name.c_str(),
      addr.address.c_str(),
  };
}

int RFC5322DateTime_parse(RFC5322DateTime* output, const char* input) {
  try {
    rfc5322::DateTime dt = rfc5322::ParseDateTime(input);
    output->day = dt.day;
    output->year = dt.year;
    output->month = dt.month;
    output->hour = dt.hour;
    output->min = dt.min;
    output->sec = dt.sec;
    if (dt.tzType == rfc5322::TimeZoneType::Offset) {
      output->tzType = TZ_TYPE_OFFSET;
      output->tz = dt.tzOffset;
    } else {
      output->tzType = TZ_TYPE_CODE;
      output->tz = static_cast<int>(dt.tz);
    }

    return 0;
  } catch (...) {
    return -1;
  }
}
}
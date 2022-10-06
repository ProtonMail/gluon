#pragma once

#include <string_view>

namespace rfc2047 {

  bool is_encoded(std::string_view input);

  std::string parse(std::string_view input);
}
#pragma once

#include <string>
#include <vector>
#include <string_view>

namespace rfc5322 {

struct Address {
  std::string name;
  std::string address;
};


std::vector<Address> ParseAddressList(std::string_view);


}
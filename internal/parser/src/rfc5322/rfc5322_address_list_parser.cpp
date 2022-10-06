#include "rfc5322_address_list_parser.h"
#include <RFC5322AddressListLexer.h>
#include <RFC5322AddressListParser.h>
#include <antlr4-runtime.h>

#include "error_listener.h"

#include "address_list_visitor.h"

namespace rfc5322 {

static std::vector<Address> buildAddressList(RFC5322AddressListParser::AddressListContext* addressListContext) {
  std::vector<Address> result;

  AddressListVisitor visitor;
  visitor.visit(addressListContext);

  return visitor.releaseAddresses();
}

std::vector<Address> ParseAddressList(std::string_view str) {
  antlr4::ANTLRInputStream inputStream{str};
  RFC5322AddressListLexer lexer{&inputStream};
  antlr4::CommonTokenStream tokens{&lexer};
  RFC5322AddressListParser parser{&tokens};

  lexer.removeErrorListeners();
  parser.removeErrorListeners();

  parser::ErrorListener errorListener;
  parser.addErrorListener(&errorListener);

  auto addressListContext = parser.addressList();

  if (errorListener.didError()) {
    throw std::runtime_error(errorListener.m_errorMessage.value());
  }

  return buildAddressList(addressListContext);
}

}  // namespace rfc5322

#pragma once

#include <RFC5322AddressListParserBaseVisitor.h>

namespace rfc5322 {

struct Address;

class AddressListVisitor final : public RFC5322AddressListParserBaseVisitor {
 public:

  AddressListVisitor();

  ~AddressListVisitor();

  std::any visitNameAddr(RFC5322AddressListParser::NameAddrContext *ctx) override;

  std::any visitGroup(RFC5322AddressListParser::GroupContext *ctx) override;

  std::any visitDisplayName(RFC5322AddressListParser::DisplayNameContext* ctx) override;

  std::any visitAngleAddr(RFC5322AddressListParser::AngleAddrContext *ctx) override;

  std::any visitMailbox(RFC5322AddressListParser::MailboxContext *ctx) override;

  std::any visitAtom(RFC5322AddressListParser::AtomContext *ctx) override;

  std::any visitEncodedAtom(RFC5322AddressListParser::EncodedAtomContext *ctx) override;

  std::any visitDotatom(RFC5322AddressListParser::DotatomContext *ctx) override;

  std::any visitWord(RFC5322AddressListParser::WordContext* ctx) override;

  std::any visitQuotedString(RFC5322AddressListParser::QuotedStringContext *context) override;

  std::any visitAddrSpec(RFC5322AddressListParser::AddrSpecContext *ctx) override;

  std::any visitLocalPart(RFC5322AddressListParser::LocalPartContext *ctx) override;

  std::any visitDomain(RFC5322AddressListParser::DomainContext *ctx) override;

  std::any visitDomainLiteral(RFC5322AddressListParser::DomainLiteralContext *ctx) override;

  std::any visitObsAngleAddr(RFC5322AddressListParser::ObsAngleAddrContext *ctx) override;

  std::any visitObsRoute(RFC5322AddressListParser::ObsRouteContext *ctx) override;

  std::any visitObsLocalPart(RFC5322AddressListParser::ObsLocalPartContext *ctx) override;

  std::any visitObsDomain(RFC5322AddressListParser::ObsDomainContext *ctx) override;


  const std::vector<Address>& getAddresses() const {
    return mAddressList;
  }

  std::vector<Address> releaseAddresses() {
    return std::move(mAddressList);
  }

 private:
  std::any visitChildren(antlr4::tree::ParseTree* tree) override;

  std::vector<Address> mAddressList;
};

}
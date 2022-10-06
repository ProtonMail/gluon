//
// Created by work on 29.09.22.
//

#include "address_list_visitor.h"
#include "rfc2047/rfc2047_parser.h"
#include "rfc5322_address_list_parser.h"

#include <cstdlib>

namespace rfc5322 {

struct DisplayWord {
  std::string value;
  bool spaceBefore;
};

std::string trimString(std::string str, const char ch) {
  str.erase(str.find_last_not_of(ch) + 1);
  str.erase(0, str.find_first_not_of(ch));
  return str;
}

std::string combineChildrenIntoString(antlr4::ParserRuleContext* context,
                                      RFC5322AddressListParserBaseVisitor* visitor) {
  std::string result;
  for (const auto& child : context->children) {
    auto visitResult = visitor->visit(child);
    if (visitResult.type() == typeid(std::string)) {
      result += std::any_cast<std::string>(visitResult);
    }
  }

  return result;
}

class DisplayWordVisitor final : public RFC5322AddressListParserBaseVisitor {
 public:
  std::any visitUnspaced(RFC5322AddressListParser::UnspacedContext* ctx) override {
    mDisplayWords.push_back(DisplayWord{
        ctx->getText(),
        false,
    });

    return std::any{};
  }

  std::any visitAtom(RFC5322AddressListParser::AtomContext* ctx) override {
    auto text = ctx->getText();
    bool spaceBefore = true;

    if (rfc2047::is_encoded(text)) {
      text = rfc2047::parse(text);
      spaceBefore = false;
    }

    mDisplayWords.push_back(DisplayWord{
        std::move(text),
        spaceBefore,
    });

    return std::any{};
  }

  std::any visitEncodedAtom(RFC5322AddressListParser::EncodedAtomContext* ctx) override {
    std::string result;
    for (const auto& encodedWord : ctx->EncodedAtom()) {
      result += rfc2047::parse(encodedWord->getText());
    }
    mDisplayWords.push_back(DisplayWord{
        result,
        true,
    });

    return std::any{};
  }

  std::any visitDotatom(RFC5322AddressListParser::DotatomContext* ctx) override {
    mDisplayWords.push_back(DisplayWord{
        ctx->getText(),
        true,
    });

    return std::any{};
  }

  std::any visitQuotedString(RFC5322AddressListParser::QuotedStringContext* ctx) override {
    mDisplayWords.push_back(DisplayWord{
        trimString(ctx->getText(), '"'),
        true,
    });

    return std::any{};
  }

  std::string combine() const {
    std::string result;

    if (mDisplayWords.empty()) {
      return std::string{};
    }

    result.append(mDisplayWords[0].value);

    auto prevIt = mDisplayWords.begin();
    for (auto it = prevIt + 1; it != mDisplayWords.end(); ++it) {
      if (it->spaceBefore && prevIt->spaceBefore) {
        result.push_back(' ');
      }
      result.append(it->value);
      prevIt = it;
    }

    return result;
  }

  std::any visitChildren(antlr4::tree::ParseTree* tree) override {
    for (const auto& child : tree->children) {
      child->accept(this);
    }

    return std::any{};
  }

 private:
  std::vector<DisplayWord> mDisplayWords;
};

AddressListVisitor::AddressListVisitor() {}

AddressListVisitor::~AddressListVisitor() {}

std::any AddressListVisitor::visitNameAddr(RFC5322AddressListParser::NameAddrContext* ctx) {
  Address address{};

  if (ctx->displayName() != nullptr) {
    address.name = std::any_cast<std::string>(visitDisplayName(ctx->displayName()));
  }

  auto angleResult = visitAngleAddr(ctx->angleAddr());
  if (angleResult.has_value()) {
    address.address = std::any_cast<std::string>(angleResult);
  }

  mAddressList.emplace_back(std::forward<Address>(address));

  return std::any{};
}

std::any AddressListVisitor::visitGroup(RFC5322AddressListParser::GroupContext* ctx) {
  if (auto groupList = ctx->groupList(); groupList != nullptr) {
    visitGroupList(groupList);
  }

  return std::any{};
}

std::any AddressListVisitor::visitDisplayName(RFC5322AddressListParser::DisplayNameContext* ctx) {
  DisplayWordVisitor visitor;
  visitor.visitChildren(ctx);
  return visitor.combine();
}

std::any AddressListVisitor::visitAngleAddr(RFC5322AddressListParser::AngleAddrContext* ctx) {
  if (auto addrSpec = ctx->addrSpec(); addrSpec != nullptr) {
    return visitAddrSpec(addrSpec);
  }

  if (auto obsAngleAddr = ctx->obsAngleAddr(); obsAngleAddr != nullptr) {
    return visit(ctx->obsAngleAddr());
  }

  return std::any{};
}

std::any AddressListVisitor::visitMailbox(RFC5322AddressListParser::MailboxContext* ctx) {
  if (auto namedAddress = ctx->nameAddr(); namedAddress != nullptr) {
    return visitNameAddr(namedAddress);
  }

  if (auto addrSpec = ctx->addrSpec(); addrSpec != nullptr) {
    auto address = std::any_cast<std::string>(visitAddrSpec(addrSpec));
    mAddressList.emplace_back(Address{"", address});
  }

  return std::any{};
}

std::any AddressListVisitor::visitWord(RFC5322AddressListParser::WordContext* ctx) {
  return combineChildrenIntoString(ctx, this);
}

std::any AddressListVisitor::visitQuotedString(RFC5322AddressListParser::QuotedStringContext* context) {
  return trimString(context->getText(), '"');
}

std::any AddressListVisitor::visitAddrSpec(RFC5322AddressListParser::AddrSpecContext* ctx) {
  std::string address = std::any_cast<std::string>(visitLocalPart(ctx->localPart()));
  address.push_back('@');
  address.append(std::any_cast<std::string>(visitDomain(ctx->domain())));

  if (auto port = ctx->port(); port != nullptr) {
    auto portText = port->getText();
    for (const char c : portText) {
      if (!std::isdigit(c)) {
        throw std::runtime_error("Invalid port specification");
      }
    }

    address.push_back(':');
    address.append(port->getText());
  }
  return address;
}

std::any AddressListVisitor::visitLocalPart(RFC5322AddressListParser::LocalPartContext* ctx) {
  return combineChildrenIntoString(ctx, this);
}

std::any AddressListVisitor::visitDomain(RFC5322AddressListParser::DomainContext* ctx) {
  return combineChildrenIntoString(ctx, this);
}

std::any AddressListVisitor::visitDomainLiteral(RFC5322AddressListParser::DomainLiteralContext* ctx) {
  return ctx->getText();
}

std::any AddressListVisitor::visitObsAngleAddr(RFC5322AddressListParser::ObsAngleAddrContext* ctx) {
  return visitChildren(ctx);
}

std::any AddressListVisitor::visitObsRoute(RFC5322AddressListParser::ObsRouteContext* ctx) {
  return visitChildren(ctx);
}

std::any AddressListVisitor::visitObsLocalPart(RFC5322AddressListParser::ObsLocalPartContext* ctx) {
  std::string result;
  const auto words = ctx->word();

  result.append(std::any_cast<std::string>(visitWord(words[0])));

  for (auto it = words.begin() + 1; it != words.end(); ++it) {
    result.push_back('.');
    result.append(std::any_cast<std::string>(visitWord(*it)));
  }

  return result;
}

std::any AddressListVisitor::visitObsDomain(RFC5322AddressListParser::ObsDomainContext* ctx) {
  return visitChildren(ctx);
}

std::any AddressListVisitor::visitAtom(RFC5322AddressListParser::AtomContext* ctx) {
  return ctx->getText();
}

std::any AddressListVisitor::visitEncodedAtom(RFC5322AddressListParser::EncodedAtomContext* ctx) {
  std::string result;
  for (const auto& encodedWord : ctx->EncodedAtom()) {
    result += rfc2047::parse(encodedWord->getText());
  }
  return result;
}

std::any AddressListVisitor::visitDotatom(RFC5322AddressListParser::DotatomContext* ctx) {
  return ctx->getText();
}

std::any AddressListVisitor::visitChildren(antlr4::tree::ParseTree* tree) {
  for (const auto& child : tree->children) {
    child->accept(this);
  }

  return std::any{};
}

}  // namespace rfc5322
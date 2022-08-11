#include "parser.h"
#include <antlr4-runtime.h>
#include "IMAPLexer.h"
#include "IMAPParser.h"
#include "error_listener.h"
#include "imap.pb.h"
#include "visitor.h"

namespace parser {

ParseResult parse(const std::string& input, const std::map<std::string, std::string>& literals, const std::string& del) {
  ParseResult result;
  try {
    antlr4::ANTLRInputStream inputStream{input};
    imap::IMAPLexer lexer{&inputStream};
    antlr4::CommonTokenStream tokens{&lexer};
    imap::IMAPParser parser{&tokens};

    lexer.removeErrorListeners();
    parser.removeErrorListeners();

    ErrorListener errorListener;
    parser.addErrorListener(&errorListener);

    imap::IMAPParser::CommandContext* commandCtx = parser.command();
    if (errorListener.didError()) {
      result.error = *errorListener.m_errorMessage;
    } else {
      result.command = Visitor{literals, del}.visit(commandCtx).as<proto::Command>().SerializeAsString();
    }
    if (auto tagCtx = commandCtx->tag(); tagCtx != nullptr) {
      result.tag = tagCtx->getText();
    }
  } catch (const std::exception& e) {
    result.error = e.what();
  } catch (...) {
    result.error = "Unexpected error occured";
  }
  return result;
}

}  // namespace parser

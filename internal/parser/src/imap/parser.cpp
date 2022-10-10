#include "parser.h"
#include <antlr4-runtime.h>
#include "IMAPLexer.h"
#include "IMAPParser.h"
#include "error_listener.h"
#include "imap.pb.h"
#include "visitor.h"


namespace parser {

class ParserWithTokenSkip final : public imap::IMAPParser {
  using imap::IMAPParser::IMAPParser;
  std::string consumeTokens(std::string count) override {
    const int skipCount = ssize_t(std::atoi(count.c_str()));
    if (skipCount <= 0) {
      throw std::runtime_error("invalid literal count");
    }
    const auto curIndex = ssize_t(_input->index());
    const auto nextIndex = curIndex + skipCount;
    //handle overflow
    if (nextIndex <0) {
     _input->seek(_input->size());
     return "";
    } else {
      std::string result = _input->getText(antlr4::misc::Interval(curIndex, nextIndex - 1));
      _input->seek(nextIndex);
      return result;
    }
  }
};

void parseInto(const char* input, const char delimiter, ParseResult& result) {
  result.error.clear();
  result.command.clear();
  result.tag.clear();

  try {
    antlr4::ANTLRInputStream inputStream{input};
    imap::IMAPLexer lexer{&inputStream};
    antlr4::CommonTokenStream tokens{&lexer};
    ParserWithTokenSkip parser{&tokens};

    lexer.removeErrorListeners();
    parser.removeErrorListeners();

    ErrorListener errorListener;
    parser.addErrorListener(&errorListener);

    auto visitor = Visitor{delimiter};
    imap::IMAPParser::CommandContext* commandCtx = parser.command();
    if (errorListener.didError()) {
      result.error = *errorListener.m_errorMessage;
    } else {
      auto command = std::any_cast<proto::Command>(visitor.visit(commandCtx));
      const size_t dataSize = command.ByteSizeLong();

      if (dataSize > size_t(std::numeric_limits<int>::max())) {
        throw std::runtime_error("Command size is too large");
      }

      result.command.resize(dataSize, 0);
      command.SerializeToArray(result.command.data(), dataSize);
    }
    if (auto tagCtx = commandCtx->tag(); tagCtx != nullptr) {
      result.tag = tagCtx->getText();
    }
  } catch (const std::exception& e) {
    result.error = e.what();
  } catch (...) {
    result.error = "Unexpected error occured";
  }
}

ParseResult parse(const char* input, const char delimiter) {
  ParseResult result;
  parseInto(input, delimiter, result);
  return result;
}

}  // namespace parser

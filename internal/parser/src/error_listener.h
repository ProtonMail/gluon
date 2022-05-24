#ifndef PARSER_ERROR_LISTENER_H
#define PARSER_ERROR_LISTENER_H

#include <antlr4-runtime.h>
#include <optional>

namespace parser {

class ErrorListener : public antlr4::BaseErrorListener {
 public:
  virtual void syntaxError(antlr4::Recognizer*, antlr4::Token*, size_t, size_t, const std::string&, std::exception_ptr)
      override;

  inline bool didError() const { return m_errorMessage.has_value(); }

  std::optional<std::string> m_errorMessage;
};

}  // namespace parser

#endif

#include "error_listener.h"

namespace parser {

void ErrorListener::syntaxError(antlr4::Recognizer*,
                                antlr4::Token*,
                                size_t,
                                size_t,
                                const std::string& msg,
                                std::exception_ptr) {
  m_errorMessage = msg;
}

}  // namespace parser

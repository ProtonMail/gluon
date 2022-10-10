#include "parser_capi.h"
#include "parser.h"

#include <iostream>

extern "C" {

struct IMAPParser {
   parser::ParseResult result;
};

IMAPParser* IMAPParser_new() {
  return new IMAPParser();
}

void IMAPParser_free(IMAPParser* parser) {
  delete parser;
}

int IMAPParser_parse(IMAPParser* parser, const char* input, char delimiter) {
  parser::parseInto(input, delimiter, parser->result);
  return parser->result.error.empty() ? 0 : -1;
}

const char* IMAPParser_getTag(IMAPParser* parser) {
  return !parser->result.tag.empty() ? parser->result.tag.c_str() : nullptr;
}

const char* IMAPParser_getError(IMAPParser* parser) {
  return !parser->result.error.empty() ? parser->result.error.c_str() : nullptr;
}

const void* IMAPParser_getCommandData(IMAPParser* parser) {
  return !parser->result.command.empty() ? parser->result.command.data() : nullptr;
}

int IMAPParser_getCommandSize(IMAPParser* parser) {
  return parser->result.command.size();
}

}
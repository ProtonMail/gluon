#ifndef PARSER_H
#define PARSER_H

#include <map>
#include <string>
#include <vector>

namespace parser {

// Even though the command tag is stored in the command, due to our architecture, if there is a parser
// error we can't serialize incomplete commands. On error we try to see if the tag element has been
// parsed and store it in the result so we can correctly handle this from Go.

struct ParseResult {
  // If not empty, contains the currently parsed command tag
  std::string tag;
  // Contains the protobuf serialized data
  std::vector<uint8_t> command;
  // If an error occurred this field will not be empty
  std::string error;
};

void parseInto(const char* input, const char delimiter, ParseResult& output);

ParseResult parse(const char* input, const char delimiter);

}  // namespace parser

#endif

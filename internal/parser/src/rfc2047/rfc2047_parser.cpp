#include "rfc2047_parser.h"

#include "RFC2047ParserBaseVisitor.h"
#include "RFC2047Lexer.h"
#include "error_listener.h"
#include <cpp-base64/base64.h>
#include <iconv.h>

#include <stdexcept>
#include <string_view>
#include <vector>

namespace rfc2047 {

class IConvHandle {
 public:
  IConvHandle() = default;

  void open(const std::string& charset) {
    close();

    iconv_t  newHandle = iconv_open("UTF-8//TRANSLIT//IGNORE", charset.c_str());
    if (newHandle == nullptr) {
      throw std::runtime_error("invalid or unsupported charset");
    }

    mInstance =newHandle;
  }

  std::string convert(const std::string& text) {
    std::vector<char> buffer(text.size() * 2, '\0');
    size_t inputSize = text.length();
    size_t outputSize = buffer.size();

    char* inputPtr = const_cast<char*>(text.c_str());
    char* outputPtr = buffer.data();

    const size_t convertedBytes = iconv(mInstance, &inputPtr, &inputSize, &outputPtr, &outputSize);
    if (convertedBytes == size_t(-1)) {
      switch (errno) {
        case EILSEQ:
        case EINVAL:
          throw std::runtime_error("invalid multibyte chars");
        default:
          throw std::runtime_error("unknown conversion error");
      }
    }

    return std::string(buffer.data(), buffer.size() - outputSize);
  }

  ~IConvHandle() {
    close();
  }
 private:
  void close() {
    if (mInstance != nullptr) {
      iconv_close(mInstance);
    }
  }

  iconv_t mInstance = nullptr;
};

static inline uint8_t fromHex(uint8_t b) {
  if (b >= '0' && b <= '9') {
    return b - '0';
  } else if (b >= 'A' && b <= 'F') {
    return b - 'A' + 10;
  } else if (b >= 'a' && b <= 'f') {
    return b - 'a' + 10;
  } else {
    throw std::runtime_error("invalid hex byte");
  }
}

static inline char readHexByte(char a, char b) {
  uint8_t hb = fromHex(a);
  uint8_t lb = fromHex(b);
  return char(hb << 4 | lb);
}

static std::string qDecode(std::string_view text) {
  std::string dec(text.length(), '\0');
  size_t n = 0;
  for (size_t i = 0; i < text.length(); i++) {
    auto c = text[i];
    if (c == '_') {
      dec[n] = ' ';
    } else if (c == '=') {
      if (i + 2 >= text.length()) {
        throw std::runtime_error("invalid word");
      }
      char b = readHexByte(text[i + 1], text[i + 2]);
      dec[n] = b;
      i += 2;
    } else if ((c <= '~' && c >= ' ') || c == '\n' || c == '\r' || c == '\t') {
      dec[n] = c;
    } else {
      throw std::runtime_error("invalid word");
    }
    n++;
  }

  dec.erase(dec.begin()+n, dec.end());
  return dec;
}

static inline std::string decodeText(std::string_view encoding, std::string_view text) {
  if (encoding.length() != 1) {
    throw std::runtime_error("invalid encoding value");
  }

  if (tolower(encoding[0]) == 'q') {
    return qDecode(text);
  } else if (tolower(encoding[0]) == 'b') {
    return base64_decode(text);
  } else {
    throw std::runtime_error("invalid encoding value");
  }
}

std::string parse(std::string_view input) {
  antlr4::ANTLRInputStream inputStream{input};
  RFC2047Lexer lexer{&inputStream};
  antlr4::CommonTokenStream tokens{&lexer};
  RFC2047Parser parser{&tokens};

  lexer.removeErrorListeners();
  parser.removeErrorListeners();

  parser::ErrorListener errorListener;
  parser.addErrorListener(&errorListener);

  auto encodedContext = parser.encodedWordList();

  if (errorListener.didError()) {
    throw std::runtime_error(errorListener.m_errorMessage.value());
  }

  std::string result;

  for (const auto& word: encodedContext->encodedWord()) {
    const auto encoding = word->Encoding()->getText();
    const auto charset = word->Token()->getText();
    const auto text = word->encodedText()->getText();

    const auto decodedText = decodeText(encoding, text);

    IConvHandle handle;
    handle.open(charset);
    result += handle.convert(decodedText);
  }

  return result;
}

bool is_encoded(std::string_view input) {
  return input.find_first_of("=?") == 0;
}

}  // namespace rfc5322

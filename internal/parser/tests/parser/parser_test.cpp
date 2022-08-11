#include "parser.h"
#include <google/protobuf/util/json_util.h>
#include <gtest/gtest.h>
#include <nlohmann/json.hpp>
#include <string>
#include <utility>
#include "imap.pb.h"

class ParserTest : public testing::Test {
 protected:
  struct Result {
    std::string tag;
    std::string json;
  };
  
  Result parse(std::string input, std::map<std::string, std::string> literals) {
    auto command = proto::Command{};
    parser::ParseResult parseResult = parser::parse(input + "\r\n", literals, "/");
    if (!parseResult.error.empty()) {
      return Result{
          parseResult.tag,
          "",
      };
    }

    command.ParseFromString(parseResult.command);

    auto opts = google::protobuf::util::JsonPrintOptions{};
    opts.always_print_primitive_fields = true;

    auto result = std::string{};
    google::protobuf::util::MessageToJsonString(command, &result, opts);

    return Result{parseResult.tag, nlohmann::json::parse(result).dump(2)};
  }
};

TEST_F(ParserTest, Capability) {
  auto result = parse("abcd CAPABILITY", {});

  EXPECT_EQ(result.tag, "abcd");
  EXPECT_EQ(result.json, R"({
  "capability": {}
})");
}

TEST_F(ParserTest, Logout) {
  auto result = parse("A023 LOGOUT", {});

  EXPECT_EQ(result.tag, "A023");
  EXPECT_EQ(result.json, R"({
  "logout": {}
})");
}

TEST_F(ParserTest, Noop) {
  auto result = parse("a002 NOOP", {});

  EXPECT_EQ(result.tag, "a002");
  EXPECT_EQ(result.json, R"({
  "noop": {}
})");
}

TEST_F(ParserTest, StartTLS) {
  auto result = parse("a002 STARTTLS", {});

  EXPECT_EQ(result.tag, "a002");
  EXPECT_EQ(result.json, R"({
  "startTLS": {}
})");
}

TEST_F(ParserTest, Authenticate) {
  auto result = parse("A001 AUTHENTICATE GSSAPI", {});

  EXPECT_EQ(result.tag, "A001");
  EXPECT_EQ(result.json, R"({
  "auth": {
    "data": [],
    "type": "GSSAPI"
  }
})");
}

TEST_F(ParserTest, Login) {
  auto result = parse("a001 LOGIN SMITH SESAME", {});

  EXPECT_EQ(result.tag, "a001");
  EXPECT_EQ(result.json, R"({
  "login": {
    "password": "SESAME",
    "username": "SMITH"
  }
})");
}

TEST_F(ParserTest, LoginQuoted) {
  auto result = parse(R"(a001 login "SMITH" "SESAME")", {});

  EXPECT_EQ(result.tag, "a001");
  EXPECT_EQ(result.json, R"({
  "login": {
    "password": "SESAME",
    "username": "SMITH"
  }
})");
}

TEST_F(ParserTest, LoginLiteral) {
  auto result = parse(
      "a001 login {5}\r\n00010203-0405-4607-8809-0a0b0c0d0e0f "
      "{6}\r\n00020203-0405-4607-8809-0a0b0c0d0e0f",
      {
          {"00010203-0405-4607-8809-0a0b0c0d0e0f", "SMITH"},
          {"00020203-0405-4607-8809-0a0b0c0d0e0f", "SESAME"},
      });

  EXPECT_EQ(result.tag, "a001");
  EXPECT_EQ(result.json, R"({
  "login": {
    "password": "SESAME",
    "username": "SMITH"
  }
})");
}

TEST_F(ParserTest, Select) {
  auto result = parse("A142 SELECT INBOX", {});

  EXPECT_EQ(result.tag, "A142");
  EXPECT_EQ(result.json, R"({
  "select": {
    "mailbox": "INBOX"
  }
})");
}

TEST_F(ParserTest, SelectInboxLowercase) {
  auto result = parse("A142 SELECT inbox", {});

  EXPECT_EQ(result.tag, "A142");
  EXPECT_EQ(result.json, R"({
  "select": {
    "mailbox": "INBOX"
  }
})");
}

TEST_F(ParserTest, Examine) {
  auto result = parse("A932 EXAMINE blurdybloop", {});

  EXPECT_EQ(result.tag, "A932");
  EXPECT_EQ(result.json, R"({
  "examine": {
    "mailbox": "blurdybloop"
  }
})");
}

TEST_F(ParserTest, Create) {
  auto result = parse("A003 CREATE foo/bar", {});

  EXPECT_EQ(result.tag, "A003");
  EXPECT_EQ(result.json, R"({
  "create": {
    "mailbox": "foo/bar"
  }
})");
}

TEST_F(ParserTest, CreateInboxChild) {
  auto result = parse("A003 CREATE inbox/bar", {});

  EXPECT_EQ(result.tag, "A003");
  EXPECT_EQ(result.json, R"({
  "create": {
    "mailbox": "INBOX/bar"
  }
})");
}

TEST_F(ParserTest, CreateInboxx) {
  auto result = parse("A003 CREATE inboxx", {});

  EXPECT_EQ(result.tag, "A003");
  EXPECT_EQ(result.json, R"({
  "create": {
    "mailbox": "inboxx"
  }
})");
}

TEST_F(ParserTest, Delete) {
  auto result = parse("A683 DELETE blurdybloop", {});

  EXPECT_EQ(result.tag, "A683");
  EXPECT_EQ(result.json, R"({
  "del": {
    "mailbox": "blurdybloop"
  }
})");
}

TEST_F(ParserTest, Rename) {
  auto result = parse("A683 RENAME mbox newName", {});

  EXPECT_EQ(result.tag, "A683");
  EXPECT_EQ(result.json, R"({
  "rename": {
    "mailbox": "mbox",
    "newName": "newName"
  }
})");
}

TEST_F(ParserTest, Subscribe) {
  auto result = parse("A002 SUBSCRIBE #news.comp.mail.mime", {});

  EXPECT_EQ(result.tag, "A002");
  EXPECT_EQ(result.json, R"({
  "sub": {
    "mailbox": "#news.comp.mail.mime"
  }
})");
}

TEST_F(ParserTest, Unsubscribe) {
  auto result = parse("A002 UNSUBSCRIBE #news.comp.mail.mime", {});

  EXPECT_EQ(result.tag, "A002");
  EXPECT_EQ(result.json, R"({
  "unsub": {
    "mailbox": "#news.comp.mail.mime"
  }
})");
}

TEST_F(ParserTest, List) {
  auto result = parse(R"(A101 LIST "" "")", {});

  EXPECT_EQ(result.tag, "A101");
  EXPECT_EQ(result.json, R"({
  "list": {
    "mailbox": "",
    "reference": ""
  }
})");
}

TEST_F(ParserTest, ListWithReference) {
  auto result = parse(R"(A102 LIST #news.comp.mail.misc "")", {});

  EXPECT_EQ(result.tag, "A102");
  EXPECT_EQ(result.json, R"({
  "list": {
    "mailbox": "",
    "reference": "#news.comp.mail.misc"
  }
})");
}

TEST_F(ParserTest, ListWithReferenceAndMailbox) {
  auto result = parse(R"(A202 LIST ~/Mail/ %)", {});

  EXPECT_EQ(result.tag, "A202");
  EXPECT_EQ(result.json, R"({
  "list": {
    "mailbox": "%",
    "reference": "~/Mail/"
  }
})");
}

TEST_F(ParserTest, LSUB) {
  auto result = parse(R"(A002 LSUB "#news." "comp.mail.*")", {});

  EXPECT_EQ(result.tag, "A002");
  EXPECT_EQ(result.json, R"({
  "lsub": {
    "mailbox": "comp.mail.*",
    "reference": "#news."
  }
})");
}

TEST_F(ParserTest, Status) {
  auto result = parse(R"(A042 STATUS foo (UIDNEXT MESSAGES))", {});

  EXPECT_EQ(result.tag, "A042");
  EXPECT_EQ(result.json, R"({
  "status": {
    "attributes": [
      "UIDNEXT",
      "MESSAGES"
    ],
    "mailbox": "foo"
  }
})");
}

TEST_F(ParserTest, Append) {
  auto result = parse(
      "A003 APPEND saved-messages (\\Seen) \"15-Nov-1984 13:37:01 +0730\" "
      "{23}\r\n00010203-0405-4607-8809-0a0b0c0d0e0f",
      {
          {"00010203-0405-4607-8809-0a0b0c0d0e0f", "My message body is here"},
      });

  EXPECT_EQ(result.tag, "A003");
  EXPECT_EQ(result.json, R"({
  "append": {
    "dateTime": {
      "date": {
        "day": 15,
        "month": 11,
        "year": 1984
      },
      "time": {
        "hour": 13,
        "minute": 37,
        "second": 1
      },
      "zone": {
        "hour": 7,
        "minute": 30,
        "sign": true
      }
    },
    "flags": [
      "\\Seen"
    ],
    "mailbox": "saved-messages",
    "message": "TXkgbWVzc2FnZSBib2R5IGlzIGhlcmU="
  }
})");
}

TEST_F(ParserTest, Check) {
  auto result = parse("a002 CHECK", {});

  EXPECT_EQ(result.tag, "a002");
  EXPECT_EQ(result.json, R"({
  "check": {}
})");
}

TEST_F(ParserTest, Close) {
  auto result = parse("a002 CLOSE", {});

  EXPECT_EQ(result.tag, "a002");
  EXPECT_EQ(result.json, R"({
  "close": {}
})");
}

TEST_F(ParserTest, Expunge) {
  auto result = parse("a002 EXPUNGE", {});

  EXPECT_EQ(result.tag, "a002");
  EXPECT_EQ(result.json, R"({
  "expunge": {}
})");
}

TEST_F(ParserTest, UIDExpunge) {
  auto result = parse("a002 UID EXPUNGE 1:*", {});

  EXPECT_EQ(result.tag, "a002");
  EXPECT_EQ(result.json, R"({
  "uidExpunge": {
    "sequenceSet": {
      "items": [
        {
          "range": {
            "begin": "1",
            "end": "*"
          }
        }
      ]
    }
  }
})");
}

TEST_F(ParserTest, Unselect) {
  auto result = parse("a002 UNSELECT", {});

  EXPECT_EQ(result.json, R"({
  "unselect": {}
})");
}

TEST_F(ParserTest, SearchNotFrom) {
  auto result = parse(R"(A281 SEARCH NOT FROM "Smith")", {});

  EXPECT_EQ(result.tag, "A281");
  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "",
        "field": "",
        "flag": "",
        "keyword": "SearchKWNot",
        "leftOp": {
          "children": [],
          "date": "",
          "field": "",
          "flag": "",
          "keyword": "SearchKWFrom",
          "size": 0,
          "text": "U21pdGg="
        },
        "size": 0,
        "text": ""
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchOrFrom) {
  auto result = parse(R"(A281 SEARCH OR FROM "Smith" FROM "Bob")", {});

  EXPECT_EQ(result.tag, "A281");
  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "",
        "field": "",
        "flag": "",
        "keyword": "SearchKWOr",
        "leftOp": {
          "children": [],
          "date": "",
          "field": "",
          "flag": "",
          "keyword": "SearchKWFrom",
          "size": 0,
          "text": "U21pdGg="
        },
        "rightOp": {
          "children": [],
          "date": "",
          "field": "",
          "flag": "",
          "keyword": "SearchKWFrom",
          "size": 0,
          "text": "Qm9i"
        },
        "size": 0,
        "text": ""
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchBcc) {
  auto result = parse(R"(A282 SEARCH BCC "mail@example.com")", {});

  EXPECT_EQ(result.tag, "A282");
  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "",
        "field": "",
        "flag": "",
        "keyword": "SearchKWBcc",
        "size": 0,
        "text": "bWFpbEBleGFtcGxlLmNvbQ=="
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchBefore) {
  auto result = parse(R"(A282 SEARCH BEFORE "1-Feb-1994")", {});

  EXPECT_EQ(result.tag, "A282");
  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "1-Feb-1994",
        "field": "",
        "flag": "",
        "keyword": "SearchKWBefore",
        "size": 0,
        "text": ""
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchSentBefore) {
  auto result = parse(R"(A282 SEARCH SENTBEFORE "1-Feb-1994")", {});

  EXPECT_EQ(result.tag, "A282");
  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "1-Feb-1994",
        "field": "",
        "flag": "",
        "keyword": "SearchKWSentBefore",
        "size": 0,
        "text": ""
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchOn) {
  auto result = parse(R"(A282 SEARCH ON "1-Feb-1994")", {});

  EXPECT_EQ(result.tag, "A282");
  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "1-Feb-1994",
        "field": "",
        "flag": "",
        "keyword": "SearchKWOn",
        "size": 0,
        "text": ""
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchSentOn) {
  auto result = parse(R"(A282 SEARCH SENTON "1-Feb-1994")", {});

  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "1-Feb-1994",
        "field": "",
        "flag": "",
        "keyword": "SearchKWSentOn",
        "size": 0,
        "text": ""
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchSince) {
  auto result = parse(R"(A282 SEARCH SINCE "1-Feb-1994")", {});

  EXPECT_EQ(result.tag, "A282");
  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "1-Feb-1994",
        "field": "",
        "flag": "",
        "keyword": "SearchKWSince",
        "size": 0,
        "text": ""
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchSentSince) {
  auto result = parse(R"(A282 SEARCH SENTSINCE "1-Feb-1994")", {});

  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "1-Feb-1994",
        "field": "",
        "flag": "",
        "keyword": "SearchKWSentSince",
        "size": 0,
        "text": ""
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchBody) {
  auto result = parse(R"(A282 SEARCH BODY "some body")", {});

  EXPECT_EQ(result.tag, "A282");
  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "",
        "field": "",
        "flag": "",
        "keyword": "SearchKWBody",
        "size": 0,
        "text": "c29tZSBib2R5"
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchCc) {
  auto result = parse(R"(A282 SEARCH CC "mail@example.com")", {});

  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "",
        "field": "",
        "flag": "",
        "keyword": "SearchKWCc",
        "size": 0,
        "text": "bWFpbEBleGFtcGxlLmNvbQ=="
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchFrom) {
  auto result = parse(R"(A282 SEARCH FROM "mail@example.com")", {});

  EXPECT_EQ(result.tag, "A282");
  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "",
        "field": "",
        "flag": "",
        "keyword": "SearchKWFrom",
        "size": 0,
        "text": "bWFpbEBleGFtcGxlLmNvbQ=="
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchKeyword) {
  auto result = parse(R"(A282 SEARCH KEYWORD something)", {});

  EXPECT_EQ(result.tag, "A282");
  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "",
        "field": "",
        "flag": "something",
        "keyword": "SearchKWKeyword",
        "size": 0,
        "text": ""
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchUnkeyword) {
  auto result = parse(R"(A282 SEARCH UNKEYWORD something)", {});

  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "",
        "field": "",
        "flag": "something",
        "keyword": "SearchKWUnkeyword",
        "size": 0,
        "text": ""
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchSubject) {
  auto result = parse(R"(A282 SEARCH SUBJECT "some subject")", {});

  EXPECT_EQ(result.tag, "A282");
  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "",
        "field": "",
        "flag": "",
        "keyword": "SearchKWSubject",
        "size": 0,
        "text": "c29tZSBzdWJqZWN0"
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchText) {
  auto result = parse(R"(A282 SEARCH TEXT "some text")", {});

  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "",
        "field": "",
        "flag": "",
        "keyword": "SearchKWText",
        "size": 0,
        "text": "c29tZSB0ZXh0"
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchTo) {
  auto result = parse(R"(A282 SEARCH TO "mail@example.com")", {});

  EXPECT_EQ(result.tag, "A282");
  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "",
        "field": "",
        "flag": "",
        "keyword": "SearchKWTo",
        "size": 0,
        "text": "bWFpbEBleGFtcGxlLmNvbQ=="
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchLarger) {
  auto result = parse(R"(A282 SEARCH LARGER 1234)", {});

  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "",
        "field": "",
        "flag": "",
        "keyword": "SearchKWLarger",
        "size": 1234,
        "text": ""
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchSmaller) {
  auto result = parse(R"(A282 SEARCH SMALLER 1234)", {});

  EXPECT_EQ(result.tag, "A282");
  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "",
        "field": "",
        "flag": "",
        "keyword": "SearchKWSmaller",
        "size": 1234,
        "text": ""
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchHeader) {
  auto result = parse(R"(A282 SEARCH HEADER "fieldName" "string")", {});

  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "",
        "field": "fieldName",
        "flag": "",
        "keyword": "SearchKWHeader",
        "size": 0,
        "text": "c3RyaW5n"
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchSeqSet) {
  auto result = parse(R"(A282 SEARCH 2:4,5,6:10,*)", {});

  EXPECT_EQ(result.tag, "A282");
  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "",
        "field": "",
        "flag": "",
        "keyword": "SearchKWSeqSet",
        "sequenceSet": {
          "items": [
            {
              "range": {
                "begin": "2",
                "end": "4"
              }
            },
            {
              "number": "5"
            },
            {
              "range": {
                "begin": "6",
                "end": "10"
              }
            },
            {
              "number": "*"
            }
          ]
        },
        "size": 0,
        "text": ""
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchUID) {
  auto result = parse(R"(A282 SEARCH UID 2:4,5,6:10,*)", {});

  EXPECT_EQ(result.tag, "A282");
  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "",
        "field": "",
        "flag": "",
        "keyword": "SearchKWUID",
        "sequenceSet": {
          "items": [
            {
              "range": {
                "begin": "2",
                "end": "4"
              }
            },
            {
              "number": "5"
            },
            {
              "range": {
                "begin": "6",
                "end": "10"
              }
            },
            {
              "number": "*"
            }
          ]
        },
        "size": 0,
        "text": ""
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchTextWithCharset) {
  auto result = parse(R"(A283 SEARCH CHARSET UTF-8 TEXT "some text")", {});

  EXPECT_EQ(result.tag, "A283");
  EXPECT_EQ(result.json, R"({
  "search": {
    "charset": "UTF-8",
    "keys": [
      {
        "children": [],
        "date": "",
        "field": "",
        "flag": "",
        "keyword": "SearchKWText",
        "size": 0,
        "text": "c29tZSB0ZXh0"
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchChildren) {
  auto result = parse(R"(A283 SEARCH (TEXT "some text" TEXT "some other text"))", {});

  EXPECT_EQ(result.tag, "A283");
  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [
          {
            "children": [],
            "date": "",
            "field": "",
            "flag": "",
            "keyword": "SearchKWText",
            "size": 0,
            "text": "c29tZSB0ZXh0"
          },
          {
            "children": [],
            "date": "",
            "field": "",
            "flag": "",
            "keyword": "SearchKWText",
            "size": 0,
            "text": "c29tZSBvdGhlciB0ZXh0"
          }
        ],
        "date": "",
        "field": "",
        "flag": "",
        "keyword": "SearchKWList",
        "size": 0,
        "text": ""
      }
    ]
  }
})");
}

TEST_F(ParserTest, SearchLiteralText) {
  auto result = parse("A284 SEARCH TEXT {6}\r\n00010203-0405-4607-8809-0a0b0c0d0e0f",
                      {
                          {"00010203-0405-4607-8809-0a0b0c0d0e0f", "hello!"},
                      });

  EXPECT_EQ(result.tag, "A284");
  EXPECT_EQ(result.json, R"({
  "search": {
    "keys": [
      {
        "children": [],
        "date": "",
        "field": "",
        "flag": "",
        "keyword": "SearchKWText",
        "size": 0,
        "text": "aGVsbG8h"
      }
    ]
  }
})");
}

TEST_F(ParserTest, Fetch) {
  auto result = parse("A654 FETCH 2:4 (FLAGS BODY[HEADER.FIELDS (FROM)])", {});

  EXPECT_EQ(result.tag, "A654");
  EXPECT_EQ(result.json, R"({
  "fetch": {
    "attributes": [
      {
        "keyword": "FetchKWFlags"
      },
      {
        "body": {
          "peek": false,
          "section": {
            "fields": [
              "FROM"
            ],
            "keyword": "HeaderFields",
            "parts": []
          }
        }
      }
    ],
    "sequenceSet": {
      "items": [
        {
          "range": {
            "begin": "2",
            "end": "4"
          }
        }
      ]
    }
  }
})");
}

TEST_F(ParserTest, FetchPart) {
  auto result = parse(R"(a001 FETCH 2:4 BODY.PEEK[4.2.2.HEADER.FIELDS.NOT (To From Subject)]<50.100>)", {});

  EXPECT_EQ(result.tag, "a001");
  EXPECT_EQ(result.json, R"({
  "fetch": {
    "attributes": [
      {
        "body": {
          "partial": {
            "begin": 50,
            "count": 100
          },
          "peek": true,
          "section": {
            "fields": [
              "To",
              "From",
              "Subject"
            ],
            "keyword": "HeaderFieldsNot",
            "parts": [
              4,
              2,
              2
            ]
          }
        }
      }
    ],
    "sequenceSet": {
      "items": [
        {
          "range": {
            "begin": "2",
            "end": "4"
          }
        }
      ]
    }
  }
})");
}

TEST_F(ParserTest, FetchBodyHeader) {
  auto result = parse(R"(a001 FETCH 2:4 BODY[HEADER])", {});

  EXPECT_EQ(result.tag, "a001");
  EXPECT_EQ(result.json, R"({
  "fetch": {
    "attributes": [
      {
        "body": {
          "peek": false,
          "section": {
            "fields": [],
            "keyword": "Header",
            "parts": []
          }
        }
      }
    ],
    "sequenceSet": {
      "items": [
        {
          "range": {
            "begin": "2",
            "end": "4"
          }
        }
      ]
    }
  }
})");
}

TEST_F(ParserTest, FetchBodyText) {
  auto result = parse(R"(a001 FETCH 2:4 BODY[TEXT])", {});

  EXPECT_EQ(result.tag, "a001");
  EXPECT_EQ(result.json, R"({
  "fetch": {
    "attributes": [
      {
        "body": {
          "peek": false,
          "section": {
            "fields": [],
            "keyword": "Text",
            "parts": []
          }
        }
      }
    ],
    "sequenceSet": {
      "items": [
        {
          "range": {
            "begin": "2",
            "end": "4"
          }
        }
      ]
    }
  }
})");
}

TEST_F(ParserTest, FetchMultiple) {
  auto result = parse(R"(a001 FETCH 2:4 (FLAGS INTERNALDATE RFC822.SIZE ENVELOPE BODY[]<50.100>))", {});

  EXPECT_EQ(result.tag, "a001");
  EXPECT_EQ(result.json, R"({
  "fetch": {
    "attributes": [
      {
        "keyword": "FetchKWFlags"
      },
      {
        "keyword": "FetchKWInternalDate"
      },
      {
        "keyword": "FetchKWRFC822Size"
      },
      {
        "keyword": "FetchKWEnvelope"
      },
      {
        "body": {
          "partial": {
            "begin": 50,
            "count": 100
          },
          "peek": false
        }
      }
    ],
    "sequenceSet": {
      "items": [
        {
          "range": {
            "begin": "2",
            "end": "4"
          }
        }
      ]
    }
  }
})");
}

TEST_F(ParserTest, FetchWholeBody) {
  auto result = parse(R"(a001 FETCH 2:4 (BODY[]))", {});

  EXPECT_EQ(result.tag, "a001");
  EXPECT_EQ(result.json, R"({
  "fetch": {
    "attributes": [
      {
        "body": {
          "peek": false
        }
      }
    ],
    "sequenceSet": {
      "items": [
        {
          "range": {
            "begin": "2",
            "end": "4"
          }
        }
      ]
    }
  }
})");
}

TEST_F(ParserTest, FetchAll) {
  auto result = parse("A654 FETCH 2:4 ALL", {});

  EXPECT_EQ(result.tag, "A654");
  EXPECT_EQ(result.json, R"({
  "fetch": {
    "attributes": [
      {
        "keyword": "FetchKWFlags"
      },
      {
        "keyword": "FetchKWInternalDate"
      },
      {
        "keyword": "FetchKWRFC822Size"
      },
      {
        "keyword": "FetchKWEnvelope"
      }
    ],
    "sequenceSet": {
      "items": [
        {
          "range": {
            "begin": "2",
            "end": "4"
          }
        }
      ]
    }
  }
})");
}

TEST_F(ParserTest, FetchFast) {
  auto result = parse("A654 FETCH 2:4 FAST", {});

  EXPECT_EQ(result.tag, "A654");
  EXPECT_EQ(result.json, R"({
  "fetch": {
    "attributes": [
      {
        "keyword": "FetchKWFlags"
      },
      {
        "keyword": "FetchKWInternalDate"
      },
      {
        "keyword": "FetchKWRFC822Size"
      }
    ],
    "sequenceSet": {
      "items": [
        {
          "range": {
            "begin": "2",
            "end": "4"
          }
        }
      ]
    }
  }
})");
}

TEST_F(ParserTest, FetchFull) {
  auto result = parse("A654 FETCH 2:4 FULL", {});

  EXPECT_EQ(result.tag, "A654");
  EXPECT_EQ(result.json, R"({
  "fetch": {
    "attributes": [
      {
        "keyword": "FetchKWFlags"
      },
      {
        "keyword": "FetchKWInternalDate"
      },
      {
        "keyword": "FetchKWRFC822Size"
      },
      {
        "keyword": "FetchKWEnvelope"
      },
      {
        "keyword": "FetchKWBody"
      }
    ],
    "sequenceSet": {
      "items": [
        {
          "range": {
            "begin": "2",
            "end": "4"
          }
        }
      ]
    }
  }
})");
}

TEST_F(ParserTest, Store) {
  auto result = parse(R"(A003 STORE 2:4,9:* +FLAGS (\Deleted \Seen))", {});

  EXPECT_EQ(result.tag, "A003");
  EXPECT_EQ(result.json, R"({
  "store": {
    "action": {
      "operation": "Add",
      "silent": false
    },
    "flags": [
      "\\Deleted",
      "\\Seen"
    ],
    "sequenceSet": {
      "items": [
        {
          "range": {
            "begin": "2",
            "end": "4"
          }
        },
        {
          "range": {
            "begin": "9",
            "end": "*"
          }
        }
      ]
    }
  }
})");
}

TEST_F(ParserTest, StoreSpacedFlags) {
  auto result = parse(R"(A003 STORE * -FLAGS \Deleted \Seen)", {});

  EXPECT_EQ(result.tag, "A003");
  EXPECT_EQ(result.json, R"({
  "store": {
    "action": {
      "operation": "Remove",
      "silent": false
    },
    "flags": [
      "\\Deleted",
      "\\Seen"
    ],
    "sequenceSet": {
      "items": [
        {
          "number": "*"
        }
      ]
    }
  }
})");
}

TEST_F(ParserTest, StoreSpacedFlagsReplace) {
  auto result = parse(R"(A003 STORE * FLAGS \Deleted \Seen)", {});

  EXPECT_EQ(result.tag, "A003");
  EXPECT_EQ(result.json, R"({
  "store": {
    "action": {
      "operation": "Replace",
      "silent": false
    },
    "flags": [
      "\\Deleted",
      "\\Seen"
    ],
    "sequenceSet": {
      "items": [
        {
          "number": "*"
        }
      ]
    }
  }
})");
}

TEST_F(ParserTest, Copy) {
  auto result = parse("A003 COPY 2:4 MEETING", {});

  EXPECT_EQ(result.tag, "A003");
  EXPECT_EQ(result.json, R"({
  "copy": {
    "mailbox": "MEETING",
    "sequenceSet": {
      "items": [
        {
          "range": {
            "begin": "2",
            "end": "4"
          }
        }
      ]
    }
  }
})");
}

TEST_F(ParserTest, Move) {
  auto result = parse("A023 MOVE 1:* Target", {});

  EXPECT_EQ(result.tag, "A023");
  EXPECT_EQ(result.json, R"({
  "move": {
    "mailbox": "Target",
    "sequenceSet": {
      "items": [
        {
          "range": {
            "begin": "1",
            "end": "*"
          }
        }
      ]
    }
  }
})");
}

TEST_F(ParserTest, MoveInbox) {
  auto result = parse("A023 MOVE 1:* inbox", {});

  EXPECT_EQ(result.tag, "A023");
  EXPECT_EQ(result.json, R"({
  "move": {
    "mailbox": "INBOX",
    "sequenceSet": {
      "items": [
        {
          "range": {
            "begin": "1",
            "end": "*"
          }
        }
      ]
    }
  }
})");
}

TEST_F(ParserTest, MoveInboxQuoted) {
  auto result = parse("A023 MOVE 1:* \"inbox\"", {});

  EXPECT_EQ(result.tag, "A023");
  EXPECT_EQ(result.json, R"({
  "move": {
    "mailbox": "INBOX",
    "sequenceSet": {
      "items": [
        {
          "range": {
            "begin": "1",
            "end": "*"
          }
        }
      ]
    }
  }
})");
}

TEST_F(ParserTest, UIDCopy) {
  auto result = parse("a001 UID COPY 2:4 MEETING", {});

  EXPECT_EQ(result.tag, "a001");
  EXPECT_EQ(result.json, R"({
  "uid": {
    "copy": {
      "mailbox": "MEETING",
      "sequenceSet": {
        "items": [
          {
            "range": {
              "begin": "2",
              "end": "4"
            }
          }
        ]
      }
    }
  }
})");
}

TEST_F(ParserTest, UIDMove) {
  auto result = parse("a003 UID MOVE 1:* Test", {});

  EXPECT_EQ(result.tag, "a003");
  EXPECT_EQ(result.json, R"({
  "uid": {
    "move": {
      "mailbox": "Test",
      "sequenceSet": {
        "items": [
          {
            "range": {
              "begin": "1",
              "end": "*"
            }
          }
        ]
      }
    }
  }
})");
}

TEST_F(ParserTest, UIDFetch) {
  auto result = parse("A999 UID FETCH 4827313:4828442 FLAGS", {});

  EXPECT_EQ(result.tag, "A999");
  EXPECT_EQ(result.json, R"({
  "uid": {
    "fetch": {
      "attributes": [
        {
          "keyword": "FetchKWFlags"
        }
      ],
      "sequenceSet": {
        "items": [
          {
            "range": {
              "begin": "4827313",
              "end": "4828442"
            }
          }
        ]
      }
    }
  }
})");
}

TEST_F(ParserTest, DISABLED_UIDSearch) {
  auto result = parse("a001 UID SEARCH ALL", {});

  EXPECT_EQ(result.tag, "a001");
  EXPECT_EQ(result.json, R"({
  "uid": { TODO }
})");
}

TEST_F(ParserTest, UIDStore) {
  auto result = parse(R"(a001 UID STORE 2:4,5:10 FLAGS (\Deleted \Seen))", {});

  EXPECT_EQ(result.tag, "a001");
  EXPECT_EQ(result.json, R"({
  "uid": {
    "store": {
      "action": {
        "operation": "Replace",
        "silent": false
      },
      "flags": [
        "\\Deleted",
        "\\Seen"
      ],
      "sequenceSet": {
        "items": [
          {
            "range": {
              "begin": "2",
              "end": "4"
            }
          },
          {
            "range": {
              "begin": "5",
              "end": "10"
            }
          }
        ]
      }
    }
  }
})");
}

TEST_F(ParserTest, Idle) {
  auto result = parse("a002 IDLE", {});

  EXPECT_EQ(result.tag, "a002");
  EXPECT_EQ(result.json, R"({
  "idle": {}
})");
}

TEST_F(ParserTest, Done) {
  auto result = parse("DONE", {});

  EXPECT_TRUE(result.tag.empty());
  EXPECT_EQ(result.json, R"({
  "done": {}
})");
}

TEST_F(ParserTest, SyntaxErrorInvalidFetchQueryWithTag) {
  std::string input = "A002 FETCH 1 (BODY[MIME])\r\n";
  auto result = parse(input, {});
  EXPECT_EQ(result.tag, "A002");
  EXPECT_TRUE(result.json.empty());
}

TEST_F(ParserTest, SyntaxErrorWithTagRandomGibberish) {
  std::string input = "A006 RANDOMGIBBERISHTHATDOESNOTMAKEAVALIDIMAPCOMMAND\r\n";
  auto result = parse(input, {});
  EXPECT_EQ(result.tag, "A006");
  EXPECT_TRUE(result.json.empty());
}

TEST_F(ParserTest, RandomBytesDoesNotCrash) {
  const uint8_t bytes[]={
        22, 3, 1, 1, 55, 1, 0, 1, 51, 3, 3, 197, 18, 92, 146, 206, 72, 40, 181, 29, 204, 229, 121, 102,
        109, 81, 40, 172, 107, 48, 135, 230, 173, 107, 115, 13, 165, 209, 62, 110, 57, 91, 172, 32, 177, 238, 14,
        114, 255, 237, 154, 71, 205, 130, 245, 131, 54, 61, 214, 67, 38, 91, 118, 207, 164, 187, 77, 58, 55, 68,
        245, 59, 166, 194, 7, 30, 0, 62, 19, 2, 19, 3, 19, 1, 192, 44, 192, 48, 0, 159, 204, 169, 204, 168, 204, 170,
        192, 43, 192, 47, 0, 158, 192, 36, 192, 40, 0, 107, 192, 35, 192, 39, 0, 103, 192, 10
  };
  std::string input = "A006 ";
  input.append(reinterpret_cast<const char*>(&bytes[0]), sizeof(bytes));
  auto result = parse(input, {});
  EXPECT_TRUE(result.tag.empty());
  EXPECT_TRUE(result.json.empty());
}

TEST_F(ParserTest, SyntaxErrorWithoutTag) {
  std::string input = "\r\n";
  auto result = parse(input, {});
  EXPECT_TRUE(result.tag.empty());
  EXPECT_TRUE(result.json.empty());
}

TEST_F(ParserTest, IdNil) {
  auto result = parse("a002 ID NIL", {});

  EXPECT_EQ(result.tag, "a002");
  EXPECT_EQ(result.json, R"({
  "idGet": {}
})");
}

TEST_F(ParserTest, IdWithOneField) {
  auto result = parse(R"(a002 ID ("name" "foo"))", {});

  EXPECT_EQ(result.tag, "a002");
  EXPECT_EQ(result.json, R"({
  "idSet": {
    "keys": {
      "name": "foo"
    }
  }
})");
}

TEST_F(ParserTest, IdMoreThan30FieldsFails) {
  auto result = parse(
      R"(a002 ID ("bar" "foo" "bar" "foo" "bar" "foo" "bar" "foo" "bar" "foo" "bar" "foo" "bar" "foo" "bar" "foo" "bar" "foo" "bar" "foo" "bar" "foo" "bar" "foo" "bar" "foo" "bar" "foo" "bar" "foo" "bar" "foo" "bar" "foo" "bar" "foo" "bar" "foo" "bar" "foo" "bar" "foo" "bar" "foo")",
      {});

  EXPECT_EQ(result.tag, "a002");
  EXPECT_TRUE(result.json.empty());
}

TEST_F(ParserTest, IdEmptyFields) {
  auto result = parse(R"(a002 ID ())", {});

  EXPECT_EQ(result.tag, "a002");
  EXPECT_EQ(result.json, R"({
  "idSet": {
    "keys": {}
  }
})");
}

TEST_F(ParserTest, IdNilField) {
  auto result = parse(R"(a002 ID ("foo" NIL))", {});

  EXPECT_EQ(result.tag, "a002");
  EXPECT_EQ(result.json, R"({
  "idSet": {
    "keys": {
      "foo": ""
    }
  }
})");
}

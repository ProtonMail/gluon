#include <gtest/gtest.h>

#include <rfc5322/rfc5322_address_list_parser.h>
#include <rfc5322/rfc5322_parser_capi.h>

#include <memory>

struct TestInput {
  const char* input;
  std::vector<rfc5322::Address> expected;
};

void validateTest(const TestInput& test) {
  SCOPED_TRACE(test.input);
  std::vector<rfc5322::Address> addresses;
  EXPECT_NO_THROW(addresses = rfc5322::ParseAddressList(test.input));
  EXPECT_EQ(addresses.size(), test.expected.size());
  for (size_t i = 0; i < addresses.size(); i++) {
    EXPECT_EQ(addresses[i].name, test.expected[i].name);
    EXPECT_EQ(addresses[i].address, test.expected[i].address);
  }
}

TEST(AddressList, ParseSingleAddressList) {
  const TestInput inputs[] = {
      {
          "BQA <somebody@gmail.com>",
          {{"BQA", "somebody@gmail.com"}},
      },
      {
          "user@example.com",
          {{"", "user@example.com"}},
      },
      {
          "John Doe <jdoe@machine.example>",
          {{"John Doe", "jdoe@machine.example"}},
      },
      {
          "Mary Smith <mary@example.net>",
          {{
              "Mary Smith",
              "mary@example.net",
          }},
      },
      {
          "\"Joe Q. Public\" <john.q.public@example.com>",
          {{
              "Joe Q. Public",
              "john.q.public@example.com",
          }},
      },
      {
          "Mary Smith <mary@x.test>",
          {{
              "Mary Smith",
              "mary@x.test",
          }},
      },
      {
          "jdoe@example.org",
          {{
              "",
              "jdoe@example.org",
          }},
      },
      {
          "Who? <one@y.test>",
          {{
              "Who?",
              "one@y.test",
          }},
      },
      {
          "<boss@nil.test>",
          {{
              "",
              "boss@nil.test",
          }},
      },
      {
          R"("Giant; \"Big\" Box" <sysservices@example.net>)",
          {{
              R"(Giant; \"Big\" Box)",
              "sysservices@example.net",
          }},
      },
      {
          "Pete <pete@silly.example>",
          {{
              "Pete",
              "pete@silly.example",
          }},
      },
      {
          "\"Mary Smith: Personal Account\" <smith@home.example>",
          {{
              "Mary Smith: Personal Account",
              "smith@home.example",
          }},
      },
      {
          "Pete(A nice \\) chap) <pete(his account)@silly.test(his host)>",
          {{
              "Pete",
              "pete@silly.test",
          }},
      },
      {
          "Gogh Fir <gf@example.com>",
          {{
              "Gogh Fir",
              "gf@example.com",
          }},
      },
      {
          "normal name  <username@server.com>",
          {{
              "normal name",
              "username@server.com",
          }},
      },
      {
          "\"comma, name\"  <username@server.com>",
          {{
              "comma, name",
              "username@server.com",
          }},
      },
      {
          "name  <username@server.com> (ignore comment)",
          {{
              "name",
              "username@server.com",
          }},
      },
      {
          "\"Mail Robot\" <>",
          {{
              "Mail Robot",
              "",
          }},
      },
      {
          "Michal Hořejšek <hořejšek@mail.com>",
          {{
              "Michal Hořejšek",
              "hořejšek@mail.com",  // Not his real address.
          }},
      },
      {
          "First Last <user@domain.com >",
          {{
              "First Last",
              "user@domain.com",
          }},
      },
      {
          "First Last <user@domain.com. >",
          {{
              "First Last",
              "user@domain.com.",
          }},
      },
      {
          "First Last <user@domain.com.>",
          {{
              "First Last",
              "user@domain.com.",
          }},
      },
      {
          "First Last <user@domain.com:25>",
          {{
              "First Last",
              "user@domain.com:25",
          }},
      },
      {
          "First Last <user@[10.0.0.1]>",
          {{
              "First Last",
              "user@[10.0.0.1]",
          }},
      },
      {
          "<postmaster@[10.10.10.10]>",
          {{
              "",
              "postmaster@[10.10.10.10]",
          }},
      },
      {
          "First Last < user@domain.com>",
          {{
              "First Last",
              "user@domain.com",
          }},
      },
      {
          "user@domain.com,",
          {{
              "",
              "user@domain.com",
          }},
      },
      {
          "First Middle \"Last\" <user@domain.com>",
          {{
              "First Middle Last",
              "user@domain.com",
          }},
      },
      {
          "First Middle Last <user@domain.com>",
          {{
              "First Middle Last",
              "user@domain.com",
          }},
      },
      {
          "First Middle\"Last\" <user@domain.com>",
          {{
              "First Middle Last",
              "user@domain.com",
          }},
      },
      {
          "First Middle \"Last\"<user@domain.com>",
          {{
              "First Middle Last",
              "user@domain.com",
          }},
      },
      {
          R"(First "Middle" "Last" <user@domain.com>)",
          {{
              "First Middle Last",
              "user@domain.com",
          }},
      },
      {
          R"(First "Middle""Last" <user@domain.com>)",
          {{
              "First Middle Last",
              "user@domain.com",
          }},
      },
      {
          "first.last <user@domain.com>",
          {{
              "first.last",
              "user@domain.com",
          }},
      },
      {
          "first . last <user@domain.com>",
          {{
              "first.last",
              "user@domain.com",
          }},
      },
      {
          "user@domain <user@domain.com>",
          {{
              "user@domain",
              "user@domain.com",
          }},
      },
      {
          "user @ domain <user@domain.com>",
          {{
              "user@domain",
              "user@domain.com",
          }},
      },
  };

  for (const auto& test : inputs) {
    validateTest(test);
  }
}

TEST(AddressList, ParseManyAddressList) {
  const TestInput inputs[]{
      {
          R"(Alice <alice@example.com>, Bob <bob@example.com>, Eve <eve@example.com>)",
          {
              {
                  R"(Alice)",
                  R"(alice@example.com)",
              },
              {
                  R"(Bob)",
                  R"(bob@example.com)",
              },
              {
                  R"(Eve)",
                  R"(eve@example.com)",
              },
          },
      },
      {
          R"(Alice <alice@example.com>; Bob <bob@example.com>; Eve <eve@example.com>)",
          {
              {
                  R"(Alice)",
                  R"(alice@example.com)",
              },
              {
                  R"(Bob)",
                  R"(bob@example.com)",
              },
              {
                  R"(Eve)",
                  R"(eve@example.com)",
              },
          },
      },
      {
          R"(Ed Jones <c@a.test>,joe@where.test,John <jdoe@one.test>)",
          {
              {
                  R"(Ed Jones)",
                  R"(c@a.test)",
              },
              {
                  "",
                  R"(joe@where.test)",
              },
              {
                  R"(John)",
                  R"(jdoe@one.test)",
              },
          },
      },
      {
          R"(name (ignore comment)  <username@server.com>,  (Comment as name) username2@server.com)",
          {
              {
                  R"(name)",
                  R"(username@server.com)",
              },
              {
                  "",
                  R"(username2@server.com)",
              },
          },
      },
      {
          R"("normal name"  <username@server.com>, "comma, name" <address@server.com>)",
          {
              {
                  R"(normal name)",
                  R"(username@server.com)",
              },
              {
                  R"(comma, name)",
                  R"(address@server.com)",
              },
          },
      },
      {
          R"("comma, one"  <username@server.com>, "comma, two" <address@server.com>)",
          {
              {
                  R"(comma, one)",
                  R"(username@server.com)",
              },
              {
                  R"(comma, two)",
                  R"(address@server.com)",
              },
          },
      },
      {
          R"(normal name  <username@server.com>, (comment)All.(around)address@(the)server.com)",
          {
              {
                  R"(normal name)",
                  R"(username@server.com)",
              },
              {
                  "",
                  R"(All.address@server.com)",
              },
          },
      },
      {
          R"(normal name  <username@server.com>, All.("comma, in comment")address@(the)server.com)",
          {
              {
                  R"(normal name)",
                  R"(username@server.com)",
              },
              {
                  "",
                  R"(All.address@server.com)",
              },
          },
      },
  };

  for (const auto& test : inputs) {
    validateTest(test);
  }
}

TEST(AddressList, ParseGroups) {
  const TestInput inputs[]{
      {
          R"(A Group:Ed Jones <c@a.test>,joe@where.test,John <jdoe@one.test>;)",
          {
              {
                  R"(Ed Jones)",
                  R"(c@a.test)",
              },
              {
                  "",
                  R"(joe@where.test)",
              },
              {
                  R"(John)",
                  R"(jdoe@one.test)",
              },
          },
      },
      {
          R"(undisclosed recipients:;)",
          {},
      },
      {
          // We permit the group to not end in a semicolon, although as per RFC5322 it really should.
          R"(undisclosed recipients:)",
          {},
      },
      {
          R"((Empty list)(start)Hidden recipients  :(nobody(that I know))  ;)",
          {},
      },
  };

  for (const auto& test : inputs) {
    validateTest(test);
  }
}

TEST(AddressList, TestEncoding) {
  const TestInput inputs[]{
      {
          R"(=?US-ASCII?Q?Keith_Moore?= <moore@cs.utk.edu>)",
          {{
              R"(Keith Moore)",
              R"(moore@cs.utk.edu)",
          }},
      },
      {
          R"(=?ISO-8859-1?Q?Keld_J=F8rn_Simonsen?= <keld@dkuug.dk>)",
          {{
              R"(Keld Jørn Simonsen)",
              R"(keld@dkuug.dk)",
          }},
      },
      {
          R"(=?ISO-8859-1?Q?Andr=E9?= Pirard <PIRARD@vm1.ulg.ac.be>)",
          {{
              R"(André Pirard)",
              R"(PIRARD@vm1.ulg.ac.be)",
          }},
      },
      {
          R"(=?ISO-8859-1?Q?Olle_J=E4rnefors?= <ojarnef@admin.kth.se>)",
          {{
              R"(Olle Järnefors)",
              R"(ojarnef@admin.kth.se)",
          }},
      },
      {
          R"(=?ISO-8859-1?Q?Patrik_F=E4ltstr=F6m?= <paf@nada.kth.se>)",
          {{
              R"(Patrik Fältström)",
              R"(paf@nada.kth.se)",
          }},
      },
      {
          R"(Nathaniel Borenstein <nsb@thumper.bellcore.com> (=?iso-8859-8?b?7eXs+SDv4SDp7Oj08A==?=))",
          {{
              R"(Nathaniel Borenstein)",
              R"(nsb@thumper.bellcore.com)",
          }},
      },
      {
          R"(=?UTF-8?B?PEJlemUgam3DqW5hPg==?= <user@domain.com>)",
          {{
              R"(<Beze jména>)",
              R"(user@domain.com)",
          }},
      },
      {
          R"(First Middle =?utf-8?Q?Last?= <user@domain.com>)",
          {{
              R"(First Middle Last)",
              R"(user@domain.com)",
          }},
      },
      {
          R"(First Middle=?utf-8?Q?Last?= <user@domain.com>)",
          {{
              R"(First Middle=?utf-8?Q?Last?=)",
              R"(user@domain.com)",
          }},
      },
      {
          R"(First Middle =?utf-8?Q?Last?=<user@domain.com>)",
          {{
              R"(First Middle Last)",
              R"(user@domain.com)",
          }},
      },
      {
          R"(First =?utf-8?Q?Middle?= =?utf-8?Q?Last?= <user@domain.com>)",
          {{
              R"(First MiddleLast)",
              R"(user@domain.com)",
          }},
      },
      {
          R"(First =?utf-8?Q?Middle?==?utf-8?Q?Last?= <user@domain.com>)",
          {{
              R"(First MiddleLast)",
              R"(user@domain.com)",
          }},
      },
      {
          R"(First "Middle"=?utf-8?Q?Last?= <user@domain.com>)",
          {{
              R"(First Middle Last)",
              R"(user@domain.com)",
          }},
      },
      {
          R"(First "Middle" =?utf-8?Q?Last?= <user@domain.com>)",
          {{
              R"(First Middle Last)",
              R"(user@domain.com)",
          }},
      },
      {
          R"(First "Middle" =?utf-8?Q?Last?=<user@domain.com>)",
          {{
              R"(First Middle Last)",
              R"(user@domain.com)",
          }},
      },
      {
          R"(=?UTF-8?B?PEJlemUgam3DqW5hPg==?= <user@domain.com>)",
          {{
              R"(<Beze jména>)",
              R"(user@domain.com)",
          }},
      },
  };

  for (const auto& test : inputs) {
    validateTest(test);
  }
}

TEST(AddressList, Invalid) {
  const std::string_view inputs[] = {
      R"("comma, name"  <username@server.com>, another, name <address@server.com>)",
      R"(username)",
      R"(=?ISO-8859-2?Q?First_Last?= <user@domain.com>, <user@domain.com,First/AAA/BBB/CCC,>)",
      R"(user@domain...com)",
      R"(=?windows-1250?Q?Spr=E1vce_syst=E9mu?=)",
      R"("'user@domain.com.'")",
      R"(<this is not an email address>)",
  };

  for (const auto& input : inputs) {
    SCOPED_TRACE(input);
    EXPECT_ANY_THROW(rfc5322::ParseAddressList(input));
  }
}

TEST(AddressList, ValidEmailValidation) {
  const std::string_view inputs[] = {
      "test@io",
      "test@iana.org",
      "test@nominet.org.uk",
      "test@about.museum",
      "a@iana.org",
      "test.test@iana.org",
      R"(!#$%&`*+/=?^`{|}~@iana.org)",
      "123@iana.org",
      "test@123.com",
      "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghiklm@iana.org",
      "test@mason-dixon.com",
      "test@c--n.com",
      "test@xn--hxajbheg2az3al.xn--jxalpdlp",
      "xn--test@iana.org",
      "1@pm.me",
  };

  for (const auto& input : inputs) {
    SCOPED_TRACE(input);
    EXPECT_NO_THROW(rfc5322::ParseAddressList(input));
  }
}

TEST(AddressList, CAPI) {
  std::unique_ptr<RFC5322AddressList, void (*)(RFC5322AddressList*)> list(RFC5322AddressList_new(),
                                                                          RFC5322AddressList_free);

  {
    const TestInput validTest = {
        R"(name (ignore comment)  <username@server.com>,  (Comment as name) username2@server.com, BQA <somebody@gmail.com>)",
        {
            {
                R"(name)",
                R"(username@server.com)",
            },
            {
                "",
                R"(username2@server.com)",
            },
            {
                "BQA",
                "somebody@gmail.com",
            },
        }};

    const int result = RFC5322AddressList_parse(list.get(), validTest.input);
    EXPECT_EQ(result, 3);
    EXPECT_EQ(nullptr, RFC5322AddressList_error_str(list.get()));
    {
      auto address = RFC5322AddressList_get(list.get(), 0);
      EXPECT_EQ(address.name, validTest.expected[0].name);
      EXPECT_EQ(address.address, validTest.expected[0].address);
    }
    {
      auto address = RFC5322AddressList_get(list.get(), 1);
      EXPECT_EQ(address.name, validTest.expected[1].name);
      EXPECT_EQ(address.address, validTest.expected[1].address);
    }
    {
      auto address = RFC5322AddressList_get(list.get(), 2);
      EXPECT_EQ(address.name, validTest.expected[2].name);
      EXPECT_EQ(address.address, validTest.expected[2].address);
    }
    {
      auto address = RFC5322AddressList_get(list.get(), 3);
      EXPECT_EQ(address.name, nullptr);
      EXPECT_EQ(address.address, nullptr);
    }
  }

  {
    const char* invalidInput = R"("comma, name"  <username@server.com>, another, name <address@server.com>)";
    const int result = RFC5322AddressList_parse(list.get(), invalidInput);
    EXPECT_EQ(result, -1);
    EXPECT_NE(0, strlen(RFC5322AddressList_error_str(list.get())));
    {
      auto address = RFC5322AddressList_get(list.get(), 0);
      EXPECT_EQ(address.name, nullptr);
      EXPECT_EQ(address.address, nullptr);
    }
  }
}

TEST(AddressList, Emoji) {
  const TestInput input = {
      R"(=?utf-8?q?Goce_Test_=F0=9F=A4=A6=F0=9F=8F=BB=E2=99=82=F0=9F=99=88?= =?utf-8?q?=F0=9F=8C=B2=E2=98=98=F0=9F=8C=B4?= <foo@bar.com>, "Proton GMX Edit" <z@bar.com>, "beta@bar.com" <beta@bar.com>, "testios12" <random@bar.com>, "random@bar.com" <random@bar.com>, =?utf-8?q?=C3=9C=C3=A4=C3=B6_Jakdij?= <another@bar.com>, =?utf-8?q?Q=C3=A4_T=C3=B6=C3=BCst_12_Edit?= <random2@bar.com>, =?utf-8?q?=E2=98=98=EF=B8=8F=F0=9F=8C=B2=F0=9F=8C=B4=F0=9F=99=82=E2=98=BA?= =?utf-8?q?=EF=B8=8F=F0=9F=98=83?= <dust@bar.com>, "Somebody Outlook" <hotmal@bar.com>)",
      {{"Goce Test 🤦🏻♂🙈🌲☘🌴", "foo@bar.com"},
       {"Proton GMX Edit", "z@bar.com"},
       {"beta@bar.com", "beta@bar.com"},
       {"testios12", "random@bar.com"},
       {"random@bar.com", "random@bar.com"},
       {"Üäö Jakdij", "another@bar.com"},
       {"Qä Töüst 12 Edit", "random2@bar.com"},
       {"☘️🌲🌴🙂☺️😃", "dust@bar.com"},
       {"Somebody Outlook", "hotmal@bar.com"}}};

  validateTest(input);
}
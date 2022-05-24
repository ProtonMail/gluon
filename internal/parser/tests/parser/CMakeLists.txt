find_package(GTest REQUIRED)
find_package(nlohmann_json REQUIRED)

add_executable(parser_test
  parser_test.cpp
)

target_link_libraries(parser_test
  parser

  GTest::GTest
  GTest::Main
  nlohmann_json::nlohmann_json
)

add_test(
  NAME parser_test
  COMMAND parser_test
)

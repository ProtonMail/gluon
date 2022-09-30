
if (WIN32)
find_package(Iconv CONFIG REQUIRED)
endif()

find_path(CPP_BASE64_INCLUDE_DIRS "cpp-base64/base64.cpp")

add_library(parser STATIC
  error_listener.h
  error_listener.cpp
  imap/parser.h
  imap/parser.cpp
  imap/visitor.h
  imap/visitor.cpp
        rfc5322/rfc5322_address_list_parser.cpp
        rfc5322/rfc5322_address_list_parser.h
        rfc5322/address_list_visitor.cpp
        rfc5322/address_list_visitor.h
        rfc2047/rfc2047_parser.cpp
        rfc2047/rfc2047_parser.h
  rfc5322/rfc5322_parser_capi.h
  rfc5322/rfc5322_parser_capi.cpp
  "${CPP_BASE64_INCLUDE_DIRS}/cpp-base64/base64.cpp"
        rfc5322/rfc5322_date_time_parser.cpp rfc5322/rfc5322_date_time_parser.h)

target_link_libraries(parser
  PUBLIC parser_gen proto
)

if (WIN32)
  target_link_libraries(parser
    PRIVATE Iconv::Charset
  )

endif()

target_include_directories(parser
  PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}
  PRIVATE ${CPP_BASE64_INCLUDE_DIRS}
)

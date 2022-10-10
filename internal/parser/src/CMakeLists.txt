
find_package(Iconv REQUIRED)

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
  rfc5322/rfc5322_date_time_parser.cpp rfc5322/rfc5322_date_time_parser.h imap/parser_capi.cpp imap/parser_capi.h)

target_link_libraries(parser
  PUBLIC parser_gen proto
)

target_link_libraries(parser
  PRIVATE Iconv::Iconv
)

target_include_directories(parser
  PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}
  PRIVATE ${CPP_BASE64_INCLUDE_DIRS}
)

if (WIN32) 
# On WIN32 we build our version of iconv with vcpkg and we need to bundle it with the libs.
add_custom_command(
  TARGET parser POST_BUILD
  COMMAND ${CMAKE_COMMAND} -E copy "${CMAKE_BINARY_DIR}/vcpkg_installed/${VCPKG_TARGET_TRIPLET}/lib/libiconv.a" ${CMAKE_ARCHIVE_OUTPUT_DIRECTORY}/libiconv.a
  COMMAND_EXPAND_LISTS
)

endif()

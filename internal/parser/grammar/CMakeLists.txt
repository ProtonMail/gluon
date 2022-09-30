set(ANTLR4_TAG 4.10.1)

set(antlr_exec_path "${CMAKE_CURRENT_BINARY_DIR}/antlr.jar")
if (NOT EXISTS ${antlr_exec_path})
  file(DOWNLOAD 
    "https://www.antlr.org/download/antlr-${ANTLR4_TAG}-complete.jar" 
    "${antlr_exec_path}"
  )
endif()
set(ANTLR4_JAR_LOCATION ${antlr_exec_path})

find_package(antlr4-runtime CONFIG REQUIRED)

find_package(antlr4-generator CONFIG REQUIRED)

antlr4_generate(IMAPLexer 
  ${CMAKE_CURRENT_SOURCE_DIR}/IMAPLexer.g4 
  LEXER 
  FALSE 
  FALSE 
  imap
)

antlr4_generate(IMAPParser
  ${CMAKE_CURRENT_SOURCE_DIR}/IMAPParser.g4
  PARSER
  FALSE
  TRUE
  imap
  "${ANTLR4_TOKEN_FILES_IMAPLexer}"
  "${ANTLR4_TOKEN_DIRECTORY_IMAPLexer}"
)

antlr4_generate(RFC5322AddressListLexer
        ${CMAKE_CURRENT_SOURCE_DIR}/RFC5322AddressListLexer.g4
        LEXER
        FALSE
        FALSE
        rfc5322
        )

antlr4_generate(RFC5322AddressListParser
        ${CMAKE_CURRENT_SOURCE_DIR}/RFC5322AddressListParser.g4
        PARSER
        FALSE
        TRUE
        rfc5322
        "${ANTLR4_TOKEN_FILES_RFC5322AddressListLexer}"
        "${ANTLR4_TOKEN_DIRECTORY_RFC5322AddressListLexer}"
        )

antlr4_generate(RFC2047Lexer
        ${CMAKE_CURRENT_SOURCE_DIR}/RFC2047Lexer.g4
        LEXER
        FALSE
        FALSE
        rfc2047
        )

antlr4_generate(RFC2047Parser
        ${CMAKE_CURRENT_SOURCE_DIR}/RFC2047Parser.g4
        PARSER
        FALSE
        TRUE
        rfc2047
        "${ANTLR4_TOKEN_FILES_RFC2047Lexer}"
        "${ANTLR4_TOKEN_DIRECTORY_RFC2047Lexer}"
        )

antlr4_generate(RFC5322DateTimeLexer
        ${CMAKE_CURRENT_SOURCE_DIR}/RFC5322DateTimeLexer.g4
        LEXER
        FALSE
        FALSE
        rfc5322
        )

antlr4_generate(RFC5322DateTimeParser
        ${CMAKE_CURRENT_SOURCE_DIR}/RFC5322DateTimeParser.g4
        PARSER
        FALSE
        TRUE
        rfc5322
        "${ANTLR4_TOKEN_FILES_RFC5322DateTimeLexer}"
        "${ANTLR4_TOKEN_DIRECTORY_RFC5322DateTimeLexer}"
        )

add_library(parser_gen STATIC
  ${ANTLR4_SRC_FILES_IMAPLexer}
  ${ANTLR4_SRC_FILES_IMAPParser}
  ${ANTLR4_SRC_FILES_RFC5322DateTimeLexer}
  ${ANTLR4_SRC_FILES_RFC5322DateTimeParser}
  ${ANTLR4_SRC_FILES_RFC5322AddressListLexer}
  ${ANTLR4_SRC_FILES_RFC5322AddressListParser}
  ${ANTLR4_SRC_FILES_RFC2047Lexer}
  ${ANTLR4_SRC_FILES_RFC2047Parser}
)

target_link_libraries(parser_gen 
  PUBLIC antlr4_static
)

target_compile_definitions(parser_gen 
  PUBLIC ANTLR4CPP_STATIC
)

target_include_directories(parser_gen 
  PUBLIC ${ANTLR4_INCLUDE_DIR}
  PUBLIC ${ANTLR4_INCLUDE_DIR_IMAPLexer}
  PUBLIC ${ANTLR4_INCLUDE_DIR_IMAPParser}
  PUBLIC ${ANTLR4_INCLUDE_DIR_RFC5322AddressListLexer}
  PUBLIC ${ANTLR4_INCLUDE_DIR_RFC5322AddressListParser}
  PUBLIC ${ANTLR4_INCLUDE_DIR_RFC5322DateTimeLexer}
  PUBLIC ${ANTLR4_INCLUDE_DIR_RFC5322DateTimeParser}
  PUBLIC ${ANTLR4_INCLUDE_DIR_RFC2047Lexer}
  PUBLIC ${ANTLR4_INCLUDE_DIR_RFC2047Parser}
)

add_custom_command(
  TARGET parser_gen POST_BUILD
  COMMAND ${CMAKE_COMMAND} -E copy $<TARGET_FILE:antlr4_static> ${CMAKE_ARCHIVE_OUTPUT_DIRECTORY}/libantlr4-runtime.a
  COMMAND_EXPAND_LISTS
)

# On Linux antlr4 has a dependency on libuuid. We need to manually copy this as well
if (UNIX AND NOT APPLE)
  find_package(unofficial-libuuid CONFIG REQUIRED)
  add_custom_command(
    TARGET parser_gen POST_BUILD
    COMMAND ${CMAKE_COMMAND} -E copy $<TARGET_FILE:unofficial::UUID::uuid> ${CMAKE_ARCHIVE_OUTPUT_DIRECTORY}/libuuid.a
    COMMAND_EXPAND_LISTS
  )
endif()



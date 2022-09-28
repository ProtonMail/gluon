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

add_library(parser_gen STATIC
  ${ANTLR4_SRC_FILES_IMAPLexer}
  ${ANTLR4_SRC_FILES_IMAPParser}
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

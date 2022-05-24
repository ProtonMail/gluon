find_package(protobuf CONFIG REQUIRED)

add_library(proto STATIC)

protobuf_generate(LANGUAGE CPP TARGET proto PROTOS imap.proto)

target_link_libraries(proto 
  PUBLIC protobuf::libprotobuf
)

target_include_directories(proto 
  PUBLIC ${CMAKE_CURRENT_BINARY_DIR}
)

add_custom_command(
  TARGET proto POST_BUILD
  COMMAND ${CMAKE_COMMAND} -E copy $<TARGET_FILE:protobuf::libprotobuf> ${CMAKE_ARCHIVE_OUTPUT_DIRECTORY}/libprotobuf.a
  COMMAND_EXPAND_LISTS
)

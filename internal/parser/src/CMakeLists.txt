add_library(parser STATIC
  error_listener.h
  error_listener.cpp
  parser.h
  parser.cpp
  visitor.h
  visitor.cpp
)

target_link_libraries(parser 
  PUBLIC parser_gen proto
)

target_include_directories(parser 
  PUBLIC ${CMAKE_CURRENT_SOURCE_DIR}
)

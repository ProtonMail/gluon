cmake_minimum_required(VERSION 3.14)

set(CMAKE_TOOLCHAIN_FILE "${CMAKE_CURRENT_SOURCE_DIR}/extern/vcpkg/scripts/buildsystems/vcpkg.cmake" CACHE STRING "toolchain")

project(Gluon VERSION 0.1.0 LANGUAGES CXX)

enable_testing()

if ("${CMAKE_SYSTEM_NAME}" STREQUAL "Windows")
  string(APPEND CMAKE_EXE_LINKER_FLAGS "-static-libgcc -static-libstdc++ -static")
endif()

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_EXPORT_COMPILE_COMMANDS YES)
set(CMAKE_CXX_EXTENSIONS OFF)
set(CMAKE_DEBUG_POSTFIX "d")

if (APPLE)
  # Can't rely only on CMake's system information as it is possible to cross-compile for both x64 and arm64 on mac OS
  if ("${VCPKG_TARGET_TRIPLET}" STREQUAL "x64-osx")
    set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}/lib/Darwin/x86_64)
  else()
    set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}/lib/Darwin/arm64)
  endif()
else()
  set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_CURRENT_SOURCE_DIR}/lib/${CMAKE_SYSTEM_NAME}/${CMAKE_SYSTEM_PROCESSOR})
endif()

add_subdirectory(grammar)
add_subdirectory(proto)
add_subdirectory(src)
add_subdirectory(tests)

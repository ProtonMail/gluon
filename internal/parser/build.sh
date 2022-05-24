#!/bin/bash

BUILD_DIR=$(mktemp -d)

check_exit() {
  if [ $? -ne 0 ]; then
    echo "Process failed: $1"
    rm -r $BUILD_DIR
    exit 1
  fi
}

git submodule update --init --recursive

cmake -B $BUILD_DIR -S . -G "Ninja" -D CMAKE_BUILD_TYPE="Release" -Wno-dev 
check_exit "Failed to configure project"

cmake --build $BUILD_DIR
check_exit "Failed to build project"

ctest --test-dir $BUILD_DIR
check_exit "Failed to run unit tests"

git submodule deinit -f --all

rm -rf $BUILD_DIR

%module parser

%include "std_string.i"
%include "std_map.i"

%{
#include "src/parser.h"
%}

%include <typemaps.i>

namespace std {
   %template(StringMap) map<string,string>;
}

%include "src/parser.h"

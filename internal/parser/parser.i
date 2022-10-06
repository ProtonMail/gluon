%module parser

%include "std_string.i"
%include "std_map.i"

%{
#include "src/imap/parser.h"
        %}

%include <typemaps.i>

namespace std {
   %template(StringMap) map<string,string>;
}

%include "src/imap/parser.h"

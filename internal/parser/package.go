package parser

//#cgo LDFLAGS: -lparser -lparser_gen -lproto -lprotobuf -lantlr4-runtime -lstdc++
//#cgo linux,amd64   LDFLAGS: -Llib/Linux/x86_64   -luuid
//#cgo darwin,amd64  LDFLAGS: -Llib/Darwin/x86_64  -framework CoreFoundation -liconv
//#cgo darwin,arm64  LDFLAGS: -Llib/Darwin/arm64  -framework CoreFoundation -liconv
//#cgo windows,amd64 LDFLAGS: -Llib/Windows/AMD64 -lOle32 -static-libgcc -static-libstdc++ -static -liconv
import "C"
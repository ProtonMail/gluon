package session

import (
	"fmt"

	"github.com/ProtonMail/gluon/internal/parser"
	"github.com/ProtonMail/gluon/internal/parser/proto"

	protobuf "google.golang.org/protobuf/proto"
)

type ParserError struct {
	error string
}

func (pe *ParserError) Error() string {
	return pe.error
}

type UnmarshalError struct {
	cmdString string
}

func (e *UnmarshalError) Error() string {
	return fmt.Sprintf("Failed to unmarshal command data: '%v'", e.cmdString)
}

// On error the first type can potentially contain tag of the command that triggered the error or empty string
// otherwise. If an error occurs the *proto.Command return result will be nil.
func parse(line []byte, literals map[string][]byte, del string) (string, *proto.Command, error) {
	lit := parser.NewStringMap()

	for k, v := range literals {
		lit.Set(k, string(v))
	}

	parseResult := parser.Parse(string(line), lit, del)

	if parseError := parseResult.GetError(); len(parseError) != 0 {
		return parseResult.GetTag(), nil, &ParserError{error: parseError}
	}

	cmd := &proto.Command{}

	if err := protobuf.Unmarshal([]byte(parseResult.GetCommand()), cmd); err != nil {
		return parseResult.GetTag(), nil, &UnmarshalError{cmdString: parseResult.GetCommand()}
	}

	return parseResult.GetTag(), cmd, nil
}

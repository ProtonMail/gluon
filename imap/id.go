package imap

import (
	"context"
	"encoding/gob"
	"fmt"
	"runtime"
	"strings"

	"github.com/ProtonMail/gluon/internal"
)

const (
	IDKeyName             = "name"
	IDKeyVersion          = "version"
	IDKeyOS               = "os"
	IdKeyOSVersion        = "os-version"
	IDKeyVendor           = "vendor"
	IDKeySupportURL       = "support-url"
	IDKeyAddress          = "address"
	IDKeyDate             = "date"
	IDKeyCommand          = "command"
	IDKeyArguments        = "arguments"
	IDKeyEnvironment      = "environment"
	IMAPIDConnMetadataKey = "rfc2971-id"
)

// ID represents the RFC 2971 IMAP ID Extension. This information can be retrieved by the connector at the context
// level. To do so please use the provided GetIMAPIDFromContext() function.
type ID struct {
	Name        string
	Version     string
	OS          string
	OSVersion   string
	Vendor      string
	SupportURL  string
	Address     string
	Date        string
	Command     string
	Arguments   string
	Environment string
	Other       map[string]string
}

func NewID() ID {
	return ID{
		Name:        "Unknown",
		Version:     "Unknown",
		OS:          "Unknown",
		OSVersion:   "Unknown",
		Vendor:      "Unknown",
		SupportURL:  "",
		Address:     "",
		Date:        "",
		Command:     "",
		Arguments:   "",
		Environment: "",
		Other:       make(map[string]string),
	}
}

func (id *ID) String() string {
	var values []string

	writeIfNotEmpty := func(key string, value string) {
		if len(value) != 0 {
			values = append(values, fmt.Sprintf(`"%v" "%v"`, key, value))
		}
	}

	writeIfNotEmpty(IDKeyName, id.Name)
	writeIfNotEmpty(IDKeyVersion, id.Version)
	writeIfNotEmpty(IDKeyOS, id.OS)
	writeIfNotEmpty(IdKeyOSVersion, id.OSVersion)
	writeIfNotEmpty(IDKeyVendor, id.Vendor)
	writeIfNotEmpty(IDKeySupportURL, id.SupportURL)
	writeIfNotEmpty(IDKeyAddress, id.Address)
	writeIfNotEmpty(IDKeyDate, id.Date)
	writeIfNotEmpty(IDKeyCommand, id.Command)
	writeIfNotEmpty(IDKeyArguments, id.Arguments)
	writeIfNotEmpty(IDKeyEnvironment, id.Environment)

	for k, v := range id.Other {
		writeIfNotEmpty(k, v)
	}

	return fmt.Sprintf("(%v)", strings.Join(values, " "))
}

func NewIMAPIDFromKeyMap(m map[string]string) ID {
	id := NewID()
	paramMap := map[string]*string{
		IDKeyName:        &id.Name,
		IDKeyVersion:     &id.Version,
		IDKeyOS:          &id.OS,
		IDKeyVendor:      &id.Vendor,
		IDKeySupportURL:  &id.SupportURL,
		IDKeyAddress:     &id.Address,
		IDKeyDate:        &id.Date,
		IDKeyCommand:     &id.Command,
		IDKeyArguments:   &id.Arguments,
		IDKeyEnvironment: &id.Environment,
	}

	for k, v := range m {
		if idv, ok := paramMap[k]; ok {
			*idv = v
		} else {
			id.Other[k] = v
		}
	}

	return id
}

func NewIMAPIDFromVersionInfo(info *internal.VersionInfo) ID {
	return ID{
		Name:       info.Name,
		Version:    info.Version.String(),
		Vendor:     info.Vendor,
		SupportURL: info.SupportURL,
		OS:         runtime.GOOS,
	}
}

func GetIMAPIDFromContext(ctx context.Context) (ID, bool) {
	if v := ctx.Value(imapIDContextKey); v != nil {
		if id, ok := v.(ID); ok {
			return id, true
		}
	}

	return ID{}, false
}

func NewContextWithIMAPID(ctx context.Context, id ID) context.Context {
	return context.WithValue(ctx, imapIDContextKey, id)
}

type imapIDContextType struct{}

var imapIDContextKey imapIDContextType

func init() {
	gob.Register(&ID{})
}

package rfc5322

import (
	"errors"
	"fmt"
	"github.com/ProtonMail/gluon/rfc822"
)

var ErrInvalidMessage = errors.New("invalid rfc5322 message")

// ValidateMessageHeaderFields checks the headers of message to verify that:
// * From and Date are present.
// * If From has multiple addresses, a Sender field must be present.
// * If Both From and Sender are present and they contain one address, they must not be equal.
func ValidateMessageHeaderFields(literal []byte) error {
	headerBytes, _ := rfc822.Split(literal)

	header, err := rfc822.NewHeader(headerBytes)
	if err != nil {
		return err
	}

	// Check for date.
	{
		value := header.Get("Date")
		if len(value) == 0 {
			return fmt.Errorf("%w: Required header field 'Date' not found or empty", ErrInvalidMessage)
		}
	}

	// Check for from.
	{
		value := header.Get("From")
		if len(value) == 0 {
			return fmt.Errorf("%w: Required header field 'From' not found or empty", ErrInvalidMessage)
		}

		// Check if From is a multi address. If so, a sender filed must be present and non-empty.
		addresses, err := ParseAddressList(value)
		if err != nil {
			return fmt.Errorf("%w: failed to parse From header: %v", ErrInvalidMessage, err)
		}

		if len(addresses) > 1 {
			senderValue := header.Get("Sender")
			if len(senderValue) == 0 {
				return fmt.Errorf("%w: Required header field 'Sender' not found or empty", ErrInvalidMessage)
			}
			_, err := ParseAddress(senderValue)
			if err != nil {
				return fmt.Errorf("%w: failed to parse Sender header: %v", ErrInvalidMessage, err)
			}
		} else {
			senderValue, ok := header.GetChecked("Sender")
			if ok {
				if len(senderValue) == 0 {
					return fmt.Errorf("%w: Required header field 'Sender' should not be empty", ErrInvalidMessage)
				}

				senderAddr, err := ParseAddress(senderValue)
				if err != nil {
					return fmt.Errorf("%w: failed to parse Sender header: %v", ErrInvalidMessage, err)
				}

				if len(senderAddr) == 1 && senderAddr[0].Address == addresses[0].Address {
					return fmt.Errorf("%w: `Sender` should not be present if equal to `From`", ErrInvalidMessage)
				}
			}
		}
	}

	return nil
}

// ValidateMessageHeaderFieldsDrafts checks the headers of message to verify that at least a valid From header is
// present.
func ValidateMessageHeaderFieldsDrafts(literal []byte) error {
	headerBytes, _ := rfc822.Split(literal)

	header, err := rfc822.NewHeader(headerBytes)
	if err != nil {
		return err
	}

	// Check for from.
	value := header.Get("From")
	if len(value) == 0 {
		return fmt.Errorf("%w: Required header field 'From' not found or empty", ErrInvalidMessage)
	}

	// Check if From is a multi address. If so, a sender filed must be present and non-empty.
	if _, err := ParseAddressList(value); err != nil {
		return fmt.Errorf("%w: failed to parse From header: %v", ErrInvalidMessage, err)
	}

	return nil
}

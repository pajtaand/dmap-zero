package utils

import (
	"errors"
	"regexp"
)

var openZitiAddressRegex = regexp.MustCompile(`connId=(\d*)\s+svcId=(\S*)\s+sourceIdentity=([\w-]+)`)

// ParseOpenZitiAddress parses a connection string to extract the connId, svcId, and sourceIdentity.
//
// The input string should follow the format: "zitiConn connId=<connId> svcId=<svcId> sourceIdentity=<sourceIdentity>".
// If the input string is correctly formatted, the extracted values are returned as strings.
// If there is an error parsing the string, an error is returned.
//
// Example:
//
//	input := "zitiConn connId=2147483649 svcId=12345 sourceIdentity=6fac665e-58fc-44ec-9918-695ef19a4c21"
//	connID, svcID, sourceIdentity, err := ParseOpenZitiAddress(input)
//	if err != nil {
//	    // handle error
//	}
func ParseOpenZitiAddress(openZitiAddress string) (string, string, string, error) {
	matches := openZitiAddressRegex.FindStringSubmatch(openZitiAddress)

	if len(matches) < 4 {
		return "", "", "", errors.New("could not find enough matches")
	}
	return matches[1], matches[2], matches[3], nil
}

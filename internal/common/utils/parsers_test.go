package utils

import (
	"testing"
)

func TestParseConnString(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedConnId string
		expectedSvcId  string
		expectedSource string
		expectErr      bool
	}{
		{
			name:           "Valid input with all fields",
			input:          "zitiConn connId=2147483649 svcId=someServiceId sourceIdentity=6fac665e-58fc-44ec-9918-695ef19a4c21",
			expectedConnId: "2147483649",
			expectedSvcId:  "someServiceId",
			expectedSource: "6fac665e-58fc-44ec-9918-695ef19a4c21",
			expectErr:      false,
		},
		{
			name:           "Valid input without svcId",
			input:          "zitiConn connId=2147483649 svcId= sourceIdentity=6fac665e-58fc-44ec-9918-695ef19a4c21",
			expectedConnId: "2147483649",
			expectedSvcId:  "",
			expectedSource: "6fac665e-58fc-44ec-9918-695ef19a4c21",
			expectErr:      false,
		},
		{
			name:           "Invalid input missing sourceIdentity",
			input:          "zitiConn connId=2147483649 svcId=someServiceId",
			expectedConnId: "",
			expectedSvcId:  "",
			expectedSource: "",
			expectErr:      true,
		},
		{
			name:           "Invalid input missing connId",
			input:          "zitiConn svcId=someServiceId sourceIdentity=6fac665e-58fc-44ec-9918-695ef19a4c21",
			expectedConnId: "",
			expectedSvcId:  "",
			expectedSource: "",
			expectErr:      true,
		},
		{
			name:           "Invalid input completely malformed",
			input:          "invalid string",
			expectedConnId: "",
			expectedSvcId:  "",
			expectedSource: "",
			expectErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connId, svcId, sourceIdentity, err := ParseOpenZitiAddress(tt.input)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}
			if connId != tt.expectedConnId {
				t.Errorf("expected connId: %s, got: %s", tt.expectedConnId, connId)
			}
			if svcId != tt.expectedSvcId {
				t.Errorf("expected svcId: %s, got: %s", tt.expectedSvcId, svcId)
			}
			if sourceIdentity != tt.expectedSource {
				t.Errorf("expected sourceIdentity: %s, got: %s", tt.expectedSource, sourceIdentity)
			}
		})
	}
}

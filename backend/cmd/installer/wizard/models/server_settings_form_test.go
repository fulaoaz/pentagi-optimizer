package models

import "testing"

func TestNormalizeMCPOriginList(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "normalizes and deduplicates", input: " HTTPS://Console.Example.com/ , https://console.example.com ", want: "https://console.example.com"},
		{name: "allows wildcard", input: "*", want: "*"},
		{name: "rejects wildcard mix", input: "*,https://console.example.com", wantErr: true},
		{name: "rejects path", input: "https://console.example.com/app", wantErr: true},
		{name: "rejects non HTTP scheme", input: "ftp://console.example.com", wantErr: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := normalizeMCPOriginList(testCase.input)
			if (err != nil) != testCase.wantErr {
				t.Fatalf("error = %v, wantErr %t", err, testCase.wantErr)
			}
			if !testCase.wantErr && got != testCase.want {
				t.Fatalf("normalized origins = %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestNormalizeMCPToolList(t *testing.T) {
	got, err := normalizeMCPToolList(" get_flow_status, list_assistants, get_flow_status ")
	if err != nil {
		t.Fatalf("normalize tools: %v", err)
	}
	if got != "get_flow_status,list_assistants" {
		t.Fatalf("normalized tools = %q", got)
	}

	if _, err := normalizeMCPToolList("get flow status"); err == nil {
		t.Fatal("invalid tool name was accepted")
	}
}

func TestValidateMCPTextRejectsLineBreaksAndOversize(t *testing.T) {
	if err := validateMCPText("valid", "字段", 5); err != nil {
		t.Fatalf("valid text rejected: %v", err)
	}
	if err := validateMCPText("line\nbreak", "字段", 100); err == nil {
		t.Fatal("line break was accepted")
	}
	if err := validateMCPText("123456", "字段", 5); err == nil {
		t.Fatal("oversize text was accepted")
	}
}

func TestMCPGovernanceValues(t *testing.T) {
	for _, value := range []string{"true", "false"} {
		if err := validateMCPBoolean(value, "布尔字段"); err != nil {
			t.Fatalf("boolean %q rejected: %v", value, err)
		}
	}
	if err := validateMCPBoolean("yes", "布尔字段"); err == nil {
		t.Fatal("invalid boolean was accepted")
	}

	for _, value := range []string{"0", "1", "100000"} {
		if _, err := parseMCPRateLimit(value); err != nil {
			t.Fatalf("rate %q rejected: %v", value, err)
		}
	}
	for _, value := range []string{"-1", "100001", "one"} {
		if _, err := parseMCPRateLimit(value); err == nil {
			t.Fatalf("invalid rate %q was accepted", value)
		}
	}

	for _, value := range []string{"scope", "destructive", "write", " DESTRUCTIVE "} {
		if _, err := parseMCPApprovalMode(value); err != nil {
			t.Fatalf("approval mode %q rejected: %v", value, err)
		}
	}
	if _, err := parseMCPApprovalMode("prompt"); err == nil {
		t.Fatal("invalid approval mode was accepted")
	}
}

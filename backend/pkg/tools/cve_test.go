package tools

import (
	"encoding/json"
	"strings"
	"testing"
)

const nvdSample = `{
  "vulnerabilities": [
    {
      "cve": {
        "id": "CVE-2021-44228",
        "published": "2021-12-10T10:15:00.000",
        "lastModified": "2023-12-10T14:00:00.000",
        "descriptions": [
          {"lang": "en", "value": "Apache Log4j2 JNDI features do not protect against attacker controlled LDAP and other JNDI related endpoints."}
        ],
        "metrics": {
          "cvssMetricV31": [
            {
              "source": "nvd@nist.gov",
              "type": "Primary",
              "cvssData": {
                "version": "3.1",
                "vectorString": "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H",
                "baseScore": 10.0,
                "baseSeverity": "CRITICAL",
                "attackVector": "NETWORK",
                "attackComplexity": "LOW",
                "privilegesRequired": "NONE",
                "userInteraction": "NONE",
                "scope": "UNCHANGED",
                "confidentialityImpact": "HIGH",
                "integrityImpact": "HIGH",
                "availabilityImpact": "HIGH"
              }
            }
          ]
        },
        "references": [
          {"url": "https://nvd.nist.gov/vuln/detail/CVE-2021-44228"},
          {"url": "https://logging.apache.org/log4j/2.x/security.html"}
        ],
        "weaknesses": [
          {
            "description": [
              {"lang": "en", "value": "CWE-502"}
            ]
          }
        ]
      }
    }
  ]
}`

func TestCVEDetailParsingAndFormatting(t *testing.T) {
	var resp cveResponse
	if err := json.Unmarshal([]byte(nvdSample), &resp); err != nil {
		t.Fatalf("failed to parse NVD sample: %v", err)
	}
	if len(resp.Vulnerabilities) != 1 {
		t.Fatalf("expected 1 vulnerability, got %d", len(resp.Vulnerabilities))
	}

	out := formatCVEDetails(resp.Vulnerabilities[0].Cve)

	checks := []string{
		"CVE-2021-44228",
		"Apache Log4j2",
		"CVSS v3.1",
		"10.0",
		"CRITICAL",
		"CWE-502",
		"https://nvd.nist.gov/vuln/detail/CVE-2021-44228",
		"https://logging.apache.org/log4j/2.x/security.html",
	}
	for _, c := range checks {
		if !strings.Contains(out, c) {
			t.Errorf("formatCVEDetails() missing %q in output:\n%s", c, out)
		}
	}
}

func TestCVEDetailNotFound(t *testing.T) {
	var resp cveResponse
	if err := json.Unmarshal([]byte(`{"vulnerabilities": []}`), &resp); err != nil {
		t.Fatalf("failed to parse empty response: %v", err)
	}
	if len(resp.Vulnerabilities) != 0 {
		t.Fatalf("expected 0 vulnerabilities")
	}
}

func TestTitleCase(t *testing.T) {
	tests := map[string]string{
		"NETWORK":   "Network",
		"network":   "Network",
		"low":       "Low",
		"":          "",
		"UNCHANGED": "Unchanged",
	}
	for in, want := range tests {
		if got := titleCase(in); got != want {
			t.Errorf("titleCase(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDedupe(t *testing.T) {
	in := []string{"CWE-502", "CWE-20", "CWE-502"}
	got := dedupe(in)
	if len(got) != 2 || got[0] != "CWE-20" || got[1] != "CWE-502" {
		t.Errorf("dedupe(%v) = %v", in, got)
	}
}

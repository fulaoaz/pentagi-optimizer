package tools

import (
	"strings"
	"testing"
)

// kevTestVulns returns a small sample of the CISA KEV catalog shape.
func kevTestVulns() []struct {
	CveID                      string `json:"cveID"`
	VendorProject              string `json:"vendorProject"`
	Product                    string `json:"product"`
	VulnerabilityName          string `json:"vulnerabilityName"`
	DateAdded                  string `json:"dateAdded"`
	ShortDescription           string `json:"shortDescription"`
	RequiredAction             string `json:"requiredAction"`
	DueDate                    string `json:"dueDate"`
	KnownRansomwareCampaignUse string `json:"knownRansomwareCampaignUse"`
	Notes                      string `json:"notes"`
} {
	return []struct {
		CveID                      string `json:"cveID"`
		VendorProject              string `json:"vendorProject"`
		Product                    string `json:"product"`
		VulnerabilityName          string `json:"vulnerabilityName"`
		DateAdded                  string `json:"dateAdded"`
		ShortDescription           string `json:"shortDescription"`
		RequiredAction             string `json:"requiredAction"`
		DueDate                    string `json:"dueDate"`
		KnownRansomwareCampaignUse string `json:"knownRansomwareCampaignUse"`
		Notes                      string `json:"notes"`
	}{
		{
			CveID:                      "CVE-2024-3400",
			VendorProject:              "Palo Alto Networks",
			Product:                    "PAN-OS",
			VulnerabilityName:          "PAN-OS Command Injection Vulnerability",
			DateAdded:                  "2024-04-12",
			ShortDescription:           "Palo Alto Networks PAN-OS contains a command injection vulnerability in the GlobalProtect feature.",
			RequiredAction:             "Apply mitigations per vendor instructions.",
			DueDate:                    "2024-05-03",
			KnownRansomwareCampaignUse: "Known",
		},
		{
			CveID:                      "CVE-2021-44228",
			VendorProject:              "Apache",
			Product:                    "Log4j",
			VulnerabilityName:          "Apache Log4j Remote Code Execution Vulnerability",
			DateAdded:                  "2021-12-10",
			ShortDescription:           "Apache Log4j2 contains a JNDI injection vulnerability.",
			RequiredAction:             "Update to patched version.",
			DueDate:                    "2021-12-31",
			KnownRansomwareCampaignUse: "Known",
			Notes:                      "Extensively exploited.",
		},
		{
			CveID:                      "CVE-2023-23397",
			VendorProject:              "Microsoft",
			Product:                    "Outlook",
			VulnerabilityName:          "Microsoft Outlook Elevation of Privilege Vulnerability",
			DateAdded:                  "2023-03-14",
			ShortDescription:           "Microsoft Outlook contains an elevation of privilege vulnerability.",
			RequiredAction:             "Apply updates.",
			DueDate:                    "2023-04-04",
			KnownRansomwareCampaignUse: "No",
		},
	}
}

func TestParseCVEList(t *testing.T) {
	got := parseCVEList("cve-2021-44228, CVE-2024-3400; cve-2023-23397")
	if len(got) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(got))
	}
	if !got["CVE-2021-44228"] || !got["CVE-2024-3400"] || !got["CVE-2023-23397"] {
		t.Fatalf("unexpected parsed set: %v", got)
	}
	if parseCVEList("") != nil {
		t.Fatalf("empty input should produce nil map")
	}
}

func TestFormatKEVResults_FilterByCVE(t *testing.T) {
	out := formatKEVResults(kevTestVulns(), "CVE-2021-44228", "", "", 10)
	if !strings.Contains(out, "CVE-2021-44228") {
		t.Fatalf("expected CVE-2021-44228 in output")
	}
	if strings.Contains(out, "CVE-2024-3400") || strings.Contains(out, "CVE-2023-23397") {
		t.Fatalf("unexpected non-matching CVE in output")
	}
	if !strings.Contains(out, "Extensively exploited") {
		t.Fatalf("expected notes section for the matched entry")
	}
}

func TestFormatKEVResults_FilterByVendor(t *testing.T) {
	out := formatKEVResults(kevTestVulns(), "", "apache", "", 10)
	if !strings.Contains(out, "Apache Log4j") || strings.Contains(out, "Palo Alto") || strings.Contains(out, "Microsoft") {
		t.Fatalf("vendor filter did not isolate Apache entries")
	}
}

func TestFormatKEVResults_FilterByProduct(t *testing.T) {
	out := formatKEVResults(kevTestVulns(), "", "", "outlook", 10)
	if !strings.Contains(out, "CVE-2023-23397") || strings.Contains(out, "CVE-2021-44228") {
		t.Fatalf("product filter did not isolate Outlook entry")
	}
}

func TestFormatKEVResults_SortedNewestFirstAndLimited(t *testing.T) {
	out := formatKEVResults(kevTestVulns(), "", "", "", 2)
	tableStart := strings.Index(out, "| CVE |")
	if tableStart < 0 {
		t.Fatalf("expected markdown table header")
	}
	table := out[tableStart:]
	newest := strings.Index(table, "CVE-2024-3400")
	second := strings.Index(table, "CVE-2023-23397")
	if newest < 0 || second < 0 || newest > second {
		t.Fatalf("expected newest entry (2024-04-12) before 2023-03-14 in table")
	}
	if strings.Contains(table, "CVE-2021-44228") {
		t.Fatalf("limit=2 should drop the third entry")
	}
}

func TestFormatKEVResults_RansomwareMarker(t *testing.T) {
	out := formatKEVResults(kevTestVulns(), "CVE-2024-3400", "", "", 10)
	if !strings.Contains(out, "| yes |") {
		t.Fatalf("expected ransomware marker 'yes' for Known ransomware use")
	}
}

func TestFormatKEVResults_NoMatch(t *testing.T) {
	out := formatKEVResults(kevTestVulns(), "CVE-2099-99999", "", "", 10)
	if !strings.Contains(out, "No known exploited vulnerabilities") {
		t.Fatalf("expected no-match message, got: %s", out)
	}
}

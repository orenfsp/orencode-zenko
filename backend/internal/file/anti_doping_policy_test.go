package file

import "testing"

func TestAntiDopingCertificatePolicy(t *testing.T) {
	policy, ok := DefaultPolicies()[PurposeAntiDopingCertificate]
	if !ok {
		t.Fatal("anti-doping certificate policy is missing")
	}
	if policy.MaxBytes != 10<<20 {
		t.Fatalf("MaxBytes = %d, want %d", policy.MaxBytes, 10<<20)
	}
	for _, contentType := range []string{"application/pdf", "image/jpeg", "image/png"} {
		if _, ok := policy.AllowedMIME[contentType]; !ok {
			t.Errorf("MIME %q is not allowed", contentType)
		}
	}
	for _, contentType := range []string{
		"application/xml",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	} {
		if _, ok := policy.AllowedMIME[contentType]; ok {
			t.Errorf("MIME %q must not be allowed", contentType)
		}
	}
}

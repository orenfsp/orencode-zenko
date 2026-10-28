package file

type Purpose string

const (
	PurposePersonalDataConsent   Purpose = "personal_data_consent"
	PurposeParentalConsent       Purpose = "parental_consent"
	PurposeAntiDopingCertificate Purpose = "anti_doping_certificate"
	PurposeFederationReport      Purpose = "federation_report"
)

type Policy struct {
	MaxBytes    int64
	AllowedMIME map[string]string
}

func DefaultPolicies() map[Purpose]Policy {
	allowedDocuments := map[string]string{
		"application/pdf": ".pdf",
		"image/jpeg":      ".jpg",
		"image/png":       ".png",
		"application/xml": ".xml",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
	}
	return map[Purpose]Policy{
		PurposePersonalDataConsent: {
			MaxBytes:    10 << 20,
			AllowedMIME: cloneMIMEPolicy(allowedDocuments),
		},
		PurposeParentalConsent: {
			MaxBytes:    10 << 20,
			AllowedMIME: cloneMIMEPolicy(allowedDocuments),
		},
		PurposeAntiDopingCertificate: {
			MaxBytes: 10 << 20,
			AllowedMIME: map[string]string{
				"application/pdf": ".pdf",
				"image/jpeg":      ".jpg",
				"image/png":       ".png",
			},
		},
		PurposeFederationReport: {
			MaxBytes: 25 << 20,
			AllowedMIME: map[string]string{
				"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": ".xlsx",
			},
		},
	}
}

func cloneMIMEPolicy(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	for contentType, extension := range source {
		result[contentType] = extension
	}
	return result
}

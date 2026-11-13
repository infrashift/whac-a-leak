package whac

rules: PRIVATE_KEY: {
	description: "PEM-encoded private key"
	severity:    "CRITICAL"
	keywords: ["private key-----"]
	multiline: true
	regex: #"-----BEGIN[ A-Z0-9]*PRIVATE KEY-----\s*((?:[A-Za-z0-9+/=]{20,}\s*){3,})-----END[ A-Z0-9]*PRIVATE KEY-----"#
	tests: {
		mustMatch: ["""
			-----BEGIN RSA PRIVATE KEY-----
			EXAMPLEKEYbodyNotARealKey0123456789abcdefghijklmn
			EXAMPLEKEYbodyNotARealKey0123456789opqrstuvwxyzAB
			EXAMPLEKEYbodyNotARealKey0123456789CDEFGHIJ+/==
			-----END RSA PRIVATE KEY-----
			"""]
		mustIgnore: [
			"""
			-----BEGIN PUBLIC KEY-----
			EXAMPLEKEYbodyNotARealKey0123456789abcdefghijklmn
			EXAMPLEKEYbodyNotARealKey0123456789opqrstuvwxyzAB
			EXAMPLEKEYbodyNotARealKey0123456789CDEFGHIJ+/==
			-----END PUBLIC KEY-----
			""",
			"""
			-----BEGIN RSA PRIVATE KEY-----
			-----END RSA PRIVATE KEY-----
			""",
		]
	}
}

rules: JWT: {
	description: "JSON Web Token"
	severity:    "MEDIUM"
	keywords: ["eyj"]
	regex: #"\b(eyJ[A-Za-z0-9_-]{10,}\.eyJ[A-Za-z0-9_-]{10,}\.[A-Za-z0-9_-]{10,})"#
	tests: {
		mustMatch: ["Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJFWEFNUExFIiwibmFtZSI6Ik5vdCBBIFJlYWwgVG9rZW4ifQ.EXAMPLEsignatureNotReal0123456789"]
		mustIgnore: [
			"header only: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		]
	}
}

package whac

import "strings"

// Samples are assembled from pieces at load time, so no complete token
// appears in this file for GitHub push protection or for whac itself.

rules: SLACK_TOKEN: {
	description: "Slack bot, user or app token"
	severity:    "CRITICAL"
	keywords: ["xoxb-", "xoxp-", "xoxa-", "xoxo-", "xoxs-", "xoxr-"]
	regex: #"\b(xox[abposr]-[0-9]{8,13}-[0-9A-Za-z-]{10,})"#
	tests: {
		mustMatch: ["SLACK_TOKEN=" + "xo" + "xb-" + "123456789012-1234567890123-EXAMPLE0a1B2c3D4e5F6g7H8"]
		mustIgnore: ["xo" + "xb-notdigits-EXAMPLE"]
	}
}

rules: SLACK_WEBHOOK_URL: {
	description: "Slack incoming webhook URL"
	severity:    "HIGH"
	keywords: ["hooks.slack.com/services/"]
	regex: #"(https://hooks\.slack\.com/services/T[A-Z0-9]{8,12}/B[A-Z0-9]{8,12}/[A-Za-z0-9]{24})"#
	tests: {
		mustMatch: ["webhook: https://hooks.slack" + ".com/services/" + "TEXAMPLE01/BEXAMPLE01/" + "EXAMPLE0a1B2c3D4e5F6g7H8"]
		mustIgnore: ["docs: https://hooks.slack.com/services/T000/B000/XXX"]
	}
}

rules: STRIPE_SECRET_KEY: {
	description: "Stripe live secret or restricted key"
	severity:    "CRITICAL"
	keywords: ["sk_live_", "rk_live_"]
	regex: #"\b((?:sk|rk)_live_[0-9A-Za-z]{24,99})\b"#
	tests: {
		mustMatch: ["STRIPE_KEY=" + "sk" + "_live_" + "EXAMPLE0a1B2c3D4e5F6g7H8i9J"]
		mustIgnore: ["STRIPE_KEY=" + "sk" + "_test_" + "EXAMPLE0a1B2c3D4e5F6g7H8i9J"]
	}
}

rules: TWILIO_API_KEY: {
	description: "Twilio API key"
	severity:    "HIGH"
	keywords: ["twilio"]
	regex: #"(?i)twilio[^\n]{0,40}?\b(SK[0-9a-fA-F]{32})\b"#
	tests: {
		mustMatch: ["TWILIO_API_KEY=" + "SK" + "0123456789abcdef" + "0123456789abcdef"]
		mustIgnore: ["TWILIO_ACCOUNT_SID=" + "AC" + "0123456789abcdef" + "0123456789abcdef"]
	}
}

rules: SENDGRID_API_KEY: {
	description: "SendGrid API key"
	severity:    "CRITICAL"
	keywords: ["sg."]
	regex: #"\b(SG\.[A-Za-z0-9_-]{22}\.[A-Za-z0-9_-]{43})(?:[^A-Za-z0-9_-]|$)"#
	tests: {
		mustMatch: ["SENDGRID_API_KEY=" + "SG" + "." + "EXAMPLE0a1B2c3D4e5F6g7" + "." + "EXAMPLE0a1B2c3D4e5F6g7H8i9J0kLmNoPqrStUvWXy"]
		mustIgnore: ["msg.send(payload)"]
	}
}

rules: SHOPIFY_TOKEN: {
	description: "Shopify access token or shared secret"
	severity:    "HIGH"
	keywords: ["shpat_", "shpca_", "shppa_", "shpss_"]
	regex: #"\b(shp(?:at|ca|pa|ss)_[a-fA-F0-9]{32})\b"#
	tests: {
		mustMatch: ["SHOPIFY_TOKEN=" + "shp" + "at_" + "0123456789abcdef" + "0123456789abcdef"]
		mustIgnore: ["shp" + "at_" + "notahexvalue"]
	}
}

rules: OPENAI_API_KEY: {
	description: "OpenAI API key"
	severity:    "CRITICAL"
	keywords: ["t3blbkfj"]
	regex: #"\b(sk-(?:proj-|svcacct-|admin-)?[A-Za-z0-9_-]{20,}T3BlbkFJ[A-Za-z0-9_-]{20,})"#
	tests: {
		mustMatch: ["OPENAI_API_KEY=" + "sk-proj-" + "EXAMPLE0a1B2c3D4e5F6g7" + "T3Blbk" + "FJ" + "EXAMPLE0a1B2c3D4e5F6g7"]
		mustIgnore: ["T3Blbk" + "FJ alone is not a key"]
	}
}

rules: ANTHROPIC_API_KEY: {
	description: "Anthropic API key"
	severity:    "CRITICAL"
	keywords: ["sk-ant-"]
	regex: #"\b(sk-ant-(?:api|admin)[0-9]{2}-[A-Za-z0-9_-]{80,})"#
	tests: {
		mustMatch: ["ANTHROPIC_API_KEY=" + "sk-" + "ant-api03-" + "EXAMPLE" + strings.Repeat("0a1B2c3D4e", 9)]
		mustIgnore: ["sk-" + "ant-api03-short"]
	}
}

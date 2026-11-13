package whac

// Allowlist entries for false positives: test fixtures and documented examples.
// Rotate a real secret. Do not allowlist it.

allowlist: fingerprints: "whac1:06dd55fb7146fa412485293d4a46af29b530ecb48a0e2a6ac7c14dc03d394f52": "fake SECRET_ASSIGNMENT sample in docs/src/content/docs/docs/guides/writing-rules.md"

allowlist: fingerprints: "whac1:18d176684f635946da970623d265c66136e1038f524d0f50b1121386df576c03": "fake URL_CREDENTIALS sample in internal/engine/filters_test.go"

allowlist: fingerprints: "whac1:1cf58a8f439f3c27e9fd16a42fe048962c2e92f933ab89b6944dc3650a870339": "fake AWS_ACCESS_KEY_ID sample in internal/engine/engine_test.go"

allowlist: fingerprints: "whac1:207e79f0d6010d8124142b3653c67deeaf031e9ab11f551e364bc602a25f9a95": "fake SECRET_ASSIGNMENT sample in internal/report/report_test.go"

allowlist: fingerprints: "whac1:407f2f946600560b34ab1e3ef00ce597454d2ca1512335c728ce058376fd95e6": "fake SECRET_ASSIGNMENT sample in internal/engine/engine_test.go"

allowlist: fingerprints: "whac1:a56a4934f08c3e0302569fb8b07fbed1df65503e34aedff5b4a00ae9d83503f4": "fake GITHUB_TOKEN sample in internal/scan/git_test.go"

allowlist: fingerprints: "whac1:ab9eb399373aecc9b685c83ce184a6e8d689d4d2545b02668facabee71248103": "fake SECRET_ASSIGNMENT sample in internal/source/git_test.go"

allowlist: fingerprints: "whac1:ceaecd45ce7e62d41df233cda112cf1e4aed182b36f61607fec15d2826878fec": "fake GITHUB_TOKEN sample in internal/cli/git_test.go"

allowlist: fingerprints: "whac1:fdd7f115a474cffe98904db0aaf79f666cfc299301a12d5b87af0318224be8da": "fake URL_CREDENTIALS sample in internal/cli/agent_test.go"

allowlist: fingerprints: "whac1:f1cc8bb06e2fa95856b21233c806c90f53a8778ab7c98d39733fc781f3d7e323": "fake GENERIC_API_KEY sample in internal/scan/agent_test.go"

allowlist: fingerprints: "whac1:1187a85299ec639adf0693abfd474622cc74ce29935bb8e296ccc5e6f1cc9d43": "fake NETRC_PASSWORD sample in internal/scan/agent_test.go"

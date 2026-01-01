# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |
| < latest| :x:                |

i recommend always using the latest version of snitch.

## Reporting a Vulnerability

if you discover a security vulnerability, please report it responsibly:

1. **do not** open a public issue for security vulnerabilities
2. email the maintainer directly or use github's [private vulnerability reporting](https://github.com/karol-broda/snitch/security/advisories/new)
3. include as much detail as possible:
   - description of the vulnerability
   - steps to reproduce
   - potential impact
   - suggested fix (if any)

## What to Expect

- acknowledgment of your report within 48 hours
- regular updates on the progress of addressing the issue
- credit in the release notes (unless you prefer to remain anonymous)

## Security Considerations

snitch reads network socket information from the operating system:

- **linux**: reads from `/proc/net/*` which requires appropriate permissions
- **macos**: uses system APIs that may require elevated privileges

snitch does not:
- make network connections (except for `snitch upgrade` which fetches from github)
- write to system files
- collect or transmit any data

## Scope

the following are considered in-scope for security reports:

- vulnerabilities in snitch code
- insecure defaults or configurations
- privilege escalation issues
- information disclosure beyond intended functionality

out of scope:

- social engineering attacks
- issues in dependencies (report to the upstream project)
- issues requiring physical access to the machine


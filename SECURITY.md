# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ---------------- |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability, please report it privately to allow time for a fix before public disclosure.

### How to Report

1. **Do NOT** create a public GitHub issue
2. Email the maintainer directly (or use GitHub's private vulnerability reporting)
3. Include as much detail as possible:
   - Description of the vulnerability
   - Steps to reproduce
   - Affected versions
   - Any potential fixes (optional)

### What to Expect

- **Response time**: Within 48 hours
- **Updates**: We will keep you informed of progress
- **Credit**: We will acknowledge your contribution (if desired)

## Security Considerations

This tool captures screenshots and window content. Consider:

- **Access control**: The server should only be accessible to trusted clients
- **Network exposure**: Avoid exposing the HTTP server to untrusted networks
- **User privacy**: Ensure the environment is appropriate for screen capture

## Best Practices

### NOT intended for untrusted environments

Given the privacy implications of screen capture, this software is designed for use **only within a controlled environment** such as:

- A local machine with trusted MCP clients (Claude Desktop, Cursor, etc.)
- An isolated virtual machine or container
- A private workstation

### For Server Package

- Run as non-root user (screenshooter-mcp)
- Use firewalld/iptables to restrict access to port 11777
- Only allow localhost (127.0.0.1) by default

### For Network Exposure (Intranet/Internet)

If you need to access this service beyond localhost (e.g., from a different machine, VM, or over a network):

1. **Use a reverse proxy** - The built-in HTTP server provides no authentication or authorization. You MUST set up a reverse proxy (nginx, Caddy, traefik, etc.) that provides:
   - TLS/HTTPS transport encryption
   - Authentication (basic auth, OAuth, etc.)
   - Rate limiting

2. **Network isolation** - Use firewall rules to limit which IPs can reach the server

3. **Consider alternatives** - For remote access, consider using SSH tunneling or a properVPN instead of exposing the service directly

### For Stdio Package

- Only use with trusted MCP clients
- Do not pipe to untrusted sources
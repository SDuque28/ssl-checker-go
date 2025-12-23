# SSL Checker 🔒

A lightweight Go CLI tool to assess SSL/TLS security configurations using the Qualys SSL Labs API.

[![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## Features ✨

- ✅ **Instant SSL/TLS security assessment** - Get detailed security grades (A+ through F)
- ✅ **Cached results** - Retrieve recent assessments without waiting
- ✅ **Progress tracking** - Real-time updates during new assessments
- ✅ **Multiple endpoints** - Detect all servers behind a domain
- ✅ **Smart recommendations** - Actionable advice based on security grade
- ✅ **Rate limiting aware** - Respects SSL Labs API limits
- ✅ **Clean output** - Well-formatted, human-readable results

## Installation 📦

### From Source
```bash
# Clone the repository
git clone https://github.com/SDuque28/ssl-checker-go.git
cd ssl-checker-go

# Build the binary
go build -o ssl-checker

# Move to your PATH (optional)
sudo mv ssl-checker /usr/local/bin/

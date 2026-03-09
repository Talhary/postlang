# Postlang

Postlang is a lighting-fast, native Windows HTTP client built with Golang and the `lxn/walk` GUI library. It serves as a lightweight alternative to Postman, offering essential API testing features with a minimal memory footprint and a sleek dark theme.

## Features
- **Native Windows UI**: Uses standard Win32 controls for maximum performance and zero Electron bloat.
- **Dark Theme**: Custom dark mode support for comfortable late-night debugging.
- **OpenAPI Import**: Instantly load and parse OpenAPI 3 (`.yaml` or `.json`) specifications to populate your workspace with available endpoints.
- **Async Execution**: Never freezes your UI while waiting for slow API responses.
- **Complete HTTP Support**: GET, POST, PUT, DELETE, PATCH, OPTIONS, and HEAD methods.
- **Custom Headers & Body**: Easily define JSON payloads and specific request headers.

## Installation
Since Postlang is a native Go application, installation is as simple as downloading the binary or compiling from source.

### Compile from Source
Ensure you have Go installed on your Windows machine.
```bash
git clone https://github.com/Talhary/postlang.git
cd postlang
go mod tidy
go build -ldflags="-H windowsgui"
```

Then run `postlang.exe`.

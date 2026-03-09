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

### Compile and Build from Source
Ensure you have Go installed on your Windows machine.
```bash
git clone https://github.com/Talhary/postlang.git
cd postlang
go mod tidy
# Generate Walk resources (required for Windows UI styles)
go get github.com/akavel/rsrc
rsrc -manifest test.manifest -o rsrc.syso
# Build the native Windows executable
go build -ldflags="-H windowsgui"
```

Then run the generated `postlang.exe`.

### Updating
To update to the latest version, simply pull the latest changes and rebuild:
```bash
git pull origin main
go build -ldflags="-H windowsgui"
```

## Contributing
Push requests (Pull Requests) are highly welcome! If you have a feature idea, bug fix, or improvement:
1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

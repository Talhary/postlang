# Postlang

Postlang is a lightning-fast, hardware-accelerated HTTP client built with Golang and the **Fyne** UI toolkit. It serves as a modern, lightweight alternative to Postman, offering essential API testing features, native dark mode support, and a responsive cross-platform interface.

## Features
- **Hardware-Accelerated UI**: Built with the Fyne toolkit for maximum performance and zero Electron bloat, rendering seamlessly across operating systems.
- **Native Dark Theme**: Automatically matches your system's dark theme for comfortable late-night debugging.
- **OpenAPI Import**: Instantly load and parse OpenAPI 3 (`.yaml` or `.json`) specifications from your local file system to populate your workspace with available endpoints.
- **Global Variables**: Define key-value pairs (e.g., `BASE_URL=https://api.example.com`) in the Variables tab and seamlessly inject them into your URLs, Headers, and JSON Bodies using the `{{KEY}}` syntax.
- **Async Execution**: Never freezes your UI while waiting for slow API responses.
- **Complete HTTP Support**: GET, POST, PUT, DELETE, PATCH, OPTIONS, and HEAD methods.
- **Custom Headers & Body**: Easily define JSON payloads and specific request headers using multiline text areas.

## Installation
Since Postlang is a native Go application leveraging Fyne, installation involves downloading the binary or compiling from source with CGO enabled.

### Compile and Build from Source
Ensure you have Go installed on your machine along with a C compiler (like GCC or MinGW) required by Fyne for graphics rendering.
```bash
git clone https://github.com/Talhary/postlang.git
cd postlang
go mod tidy

# Build the native executable (Windows example)
$env:GOARCH="amd64"
$env:CGO_ENABLED="1"
go build -ldflags="-H windowsgui" -o postlang.exe .
```

Then run the generated `postlang.exe`.

### Updating
To update to the latest version, simply pull the latest changes and rebuild:
```bash
git pull origin main
go build -ldflags="-H windowsgui" -o postlang.exe .
```

## Contributing
Push requests (Pull Requests) are highly welcome! If you have a feature idea, bug fix, or improvement:
1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

module github.com/LastStep/Bonsai

go 1.25.0

toolchain go1.25.9

// v0.1.0 has a case-insensitive file collision (station/index.md vs station/INDEX.md)
// that prevents the Go module proxy from creating a valid zip.
retract v0.1.0

require golang.org/x/sys v0.47.0

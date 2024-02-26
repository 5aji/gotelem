# GoTelem: Golang-based telemetry tools

`GoTelem` is a toolkit and library to make working with solar car telemetry
data fast, reliable, and insightful. 

Features:

- SocketCAN interface for connecting to physical hardware.
- TCP streaming system based around MessagePack-RPC for LAN control/inspection.
- XBee integration and control for long-range communication.
- HTTP API for easy external tool integration.
- SQLite database format for storing telemetry, and tools to work with it.


`GoTelem` provides a flexible system for ingesting, storing, analyzing, and distributing
telemetry information.



## Rationale

There are probably two questions:

1. What's this for?
2. Why is it written in Go?

Telemetry is an interesting system since it not only involves a microcontroller on the car acting as a transmitter,
it also requires software running on a laptop that can recieve the data and do useful things with it. 
Previous iterations of this PC software usually involved Python scripts that were thrown together quickly
due to time constraints. This has a few problems, namely that performance is usually limited,
APIs are not type-safe, and environments are not portable and require setup.

So we aught to invest in better tooling - schemas and programs that make working with 
the data we collect easier and more consistent, as well as being [the standard](https://xkcd.com/927/).
This tool/repo aims to package several ideas and utilities into a single, all-in-one binary.
While that's a noble goal, design decisions are being made to support long-term evolution
of software; we have versioned SQLite databases, that are entirely standalone.

I chose to write this in Go because Go has good concurrency support, good cross-compilation,
and relatively good performance, especially when compared to interpreted languages. 

C/C++ was eliminated due to being too close to the metal and having bad tooling/cross compilation.

Python was eliminated due to having poor concurrency support and difficult packaging/distribution.
It also lacks a good http/networking story, instead relying on WSGI/ASGI packages which
make Windows not viable. Futhermore, being dynamically typed leads to issues in asserting
robustness of the code.

Rust was elminiated due to being too different from more common programming languages. Likewise
for F#, C#, D, Zig, Nim, Julia, Racket, Elixr, and Common Lisp. Yes, I did seriouisly consider each
of these. C# was a viable competitor but had issues with the cross-platform story.

Go has some quirks and -isms, like lacking "true" Object-Orientation, but the language is designed
around being normal to look at, easy to write, and straightforward to understand.

Go has really good cross compilation support built in to the `go` binary. However, since this
package uses some C libraries (SQLite), certain functionality will be missing if you don't
have a cross-compiler set up. There's a way to make things easier on Linux
[using `zig cc`](https://zig.news/kristoff/building-sqlite-with-cgo-for-every-os-4cic)
but this is not usually important since the user can pretty easily compile it on their
own system, and it's a single executable to share to others with the same OS/architecture.

## Building

`gotelem` was designed to be all-inclusive while being easy to build and have good cross-platform support. 
Binaries are a single, statically linked file that can be shared to other users of the same OS.
Certain features, like socketCAN support, are only enabled on platforms that support them (Linux). 
This is handled automatically; builds will exclude the socketCAN files and 
the additional commands and features will not be present in the CLI.


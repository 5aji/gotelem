# GoTelem: Golang-based telemetry tools

`GoTelem` is a toolkit and library to make working with solar car telemetry
data fast, reliable, and insightful. 

Features:

- SocketCAN interface for connecting to physical hardware.
- TCP streaming system based around MessagePack-RPC for LAN control/inspection.
- XBee integration and control for long-range communication.
- HTTP API for easy external tool integration.


`GoTelem` provides a flexible system for ingesting, storing, analyzing, and distributing
telemetry information.



## Rationale

There are probably two questions:

1. Why a telemetry library that runs on an OS?
2. Why is it written in Go?

To answer the first question, the needs of the telemetry board are ill-suited for a microcontroller
since it requires doing multiple non-trivial tasks in parallel. The on-car system must ingest
all can packets, write them to disk, and then transmit them over XBee if they match a filter.
Doing fast disk I/O is difficult.

There are also significant advantages to moving to using a Linux system for telemetry. We gain
Wifi/Bluetooth/network support easily, we can integrate USB devices like a USB GPS reciever,
and we can share common tooling between the car code and the receiver code.

I chose to write this in Go because Go has good concurrency support, good cross-compilation,
and relatively good performance. 

C/C++ was eliminated due to being too close to the metal and having bad tooling/cross compilation.

Python was eliminated due to having poor concurrency support and difficult packaging/distribution.
It also lacks a good http/networking story, instead relying on WSGI/ASGI packages which
make Windows not viable. Futhermore, being dynamically typed leads to issues in asserting
robustness of the code.

Rust was elminiated due to being too different from more common programming languages. Likewise
for F#, C#, D, Zig, Nim, Julia, Racket, Elixr, and Common Lisp. Yes, I did seriouisly consider each
of these.

Go has some quirks and -isms, like lacking "true" Object-Orientation, but the language is designed
around being normal to look at, easy to write, and straightforward to understand.

Go has really good cross compilation support built in to the `go` binary. However, since this
package uses some C libraries (SQLite), certain functionality will be missing if you don't
have a cross-compiler set up. There's a way to make things easier on Linux
[using `zig cc`](https://zig.news/kristoff/building-sqlite-with-cgo-for-every-os-4cic)
but this is not usually important since the user can pretty easily compile it on their
own system, and it's a single executable to share to others with the same OS/architecture.

## Building

There are build tags to enable/disable certain features, like the graphical GUI.
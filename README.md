[![GoDoc](https://godoc.org/github.com/auroralaboratories/pulse?status.svg)](https://godoc.org/github.com/auroralaboratories/pulse)

## Pulse

Pulse is a Golang wrapper around the PulseAudio C client library `libpulse`.  This library can be used to provide audio playback, mixing, and control capabilities your programs using the PulseAudio sound server on Linux and other supported systems.

## Prerequisites

- PulseAudio client and development libraries (`libpulse-dev` [Ubuntu, Debian] or `libpulse-devel` [RedHat, CentOS]).
- Golang >= 1.12

## Usage

A command-line utility, [`pulse`](cmd/pulse) is provided as a reference implementation of the library.  This code can be used as a real-world example of how to use different features of this package, as well as being a useful standalone tool for working with PulseAudio.
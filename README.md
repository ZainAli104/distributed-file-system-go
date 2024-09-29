# Peer-to-peer content-addressable distributed file system

## Overview

A Content Addressable File System (CADFS) is a type of file system that uses content-based addressing to manage and retrieve data across a network of nodes. Unlike traditional file systems that use hierarchical paths or filenames to locate data, a CADFS identifies and accesses files based on their content. This project offers a peer-to-peer content-addressable file storage built using Go. It also implements a peer-to-peer library built on top of TCP from scratch. This custom network library supports streaming files across the network in chunks, allowing exchange of large files across the network.

## Key features

- Encryption and decryption during data storage and transmission
- Content addressable storage
- Distributed storage
- Data redundancy to ensure fault tolerance
- Data streaming support to send files in chunks for exchanging large files through the network

## Architecture

![architecture of the file system](https://raw.githubusercontent.com/priyangshupal/documentation-images/main/cas-distributed-file-system/architecture.svg)

## Getting started

### Prerequisites

It is recommended to have Go installed before running the project. Go can be installed from the official [Go website](https://go.dev/doc/install). This project was built using Go version `1.22.2`.

### Steps to run

The project dependencies need to be in place for running the project. To install the dependencies, inside the project directory, run:

```go
go mod tidy
```

Once the project dependencies are installed, the project can be run using:

```go
make run
```

This will spin up three peers and the peer started on port 4000 will create five files and broadcast them to the other peers for redundancy. If successfully run, you will see three folders named `:3000_network`, `:4000_network`, and `:5000_network` in the root of the project directory.

---
layout: default
title: Installation
nav_order: 2
permalink: /install
---

# Installation
{: .no_toc }

## Table of Contents
{: .no_toc .text-delta }

- TOC
{:toc}

---

## Installation Methods

There are several ways to install **jsonschemaprofiles**:

### Download Binary

Download pre-built binaries from the [GitHub Releases](https://github.com/UnitVectorY-Labs/jsonschemaprofiles/releases) page for the latest version.

[![GitHub release](https://img.shields.io/github/release/UnitVectorY-Labs/jsonschemaprofiles.svg)](https://github.com/UnitVectorY-Labs/jsonschemaprofiles/releases/latest) 

Choose the appropriate binary for your platform and add it to your PATH.

### Install Using Go

Install directly from the Go toolchain:

```bash
go install github.com/UnitVectorY-Labs/jsonschemaprofiles/cmd/jsonschemaprofiles@latest
```

This installs the command-line tool globally, allowing you to run `jsonschemaprofiles` from any terminal. The intended use of this is as a library, but it can be used as a CLI tool as well for quick testing and development.

### Build from Source

Build the application from source code:

```bash
git clone https://github.com/UnitVectorY-Labs/jsonschemaprofiles.git
cd jsonschemaprofiles
go build -o jsonschemaprofiles ./cmd/jsonschemaprofiles/
```

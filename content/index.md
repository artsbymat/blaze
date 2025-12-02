---
title: Welcome to Blaze
---

# Welcome to Blaze SSG

Blaze is a **blazing fast** Static Site Generator built with Go, featuring hot reload capabilities for an amazing development experience.

## Features

- Fast markdown parsing
- Template-based rendering
- Hot reload with live browser refresh
- Simple and intuitive CLI

## Getting Started

To build your site:

```
go run cmd/ssg/main.go build
```

To serve with hot reload:

```
go run cmd/ssg/main.go serve
```

## How It Works

Blaze watches your `blaze.config.json`, `content` and `templates` directories for changes. When you save a file, it automatically rebuilds and refreshes your browser.

### Example Code

Here's a simple example of _inline code_: `fmt.Println("Hello, Blaze!")`

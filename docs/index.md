# Introduction

n8go-docs is a powerful clean documentation generator written in Golang.

## Features

- Write documentation using Markdown
- Outputs static HTML that can be hosted anywhere (Github Pages, S3, etc)
- Supports syntax highlighting and emojis
- Easily extensible with custom Jinja2-compatible themes
- In-built search (powered by [FlexSearch](https://github.com/nextapps-de/flexsearch) or [Fuse.js](https://fusejs.io/))
- Supports `site_url`, `site_name`, `site_description`, `dev_addr`, and `use_directory_urls`



## Installation
- Head over to the [releases](https://github.com/maksyutin/n8go-docs/releases) page and download the latest binary for your platform.

## Usage

### Create a new project

```bash
git clone https://github.com/maksyutin/n8go-docs.git
```

### Start the development server

```bash
n8go-docs serve
```

### Build the static site

```bash
n8go-docs generate
```

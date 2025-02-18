# Krems

[A little blog that explains Krems](https://mreider.com/tech/krems-static-blog-generator/).

**Krems** is a simple, lightweight static site generator written in Go. It converts Markdown files into responsive HTML pages, complete with navigation menus, optional blog listings, and an RSS feed. Think of it like Hugo or Jekyll, but stripped down to the essentials.

## Features

- Minimal configuration: a single `config.yaml` for site name, URL, and menu structure
- Friendly “front matter” in each Markdown file for title, description, etc.
- Automatic generation of “list” pages that group sub-pages by date
- Responsive output using Bootstrap (bundled automatically)
- Integrated local server (`--run`) for quick previews
- 404 page generation for GitHub Pages or other static hosts
- RSS feed for dated pages
- Automatic CNAME file created for Github pages

---

## Installation

### Download Prebuilt Binaries

Precompiled binaries for multiple platforms are available on the [Krems GitHub Releases page](https://github.com/mreider/krems/releases). Look for files named like:

- `krems-darwin-amd64`
- `krems-darwin-arm64`
- `krems-linux-amd64`
- `krems-linux-arm64`
- `krems-windows-amd64.exe`
  
Pick the one that matches your environment. For example:

#### macOS (Intel)

```bash
curl -L https://github.com/mreider/krems/releases/download/vX.X.X/krems-darwin-amd64 -o krems
chmod +x krems
```

#### macOS (Apple Silicon / M1 / M2)

```bash
curl -L https://github.com/mreider/krems/releases/download/vX.X.X/krems-darwin-arm64 -o krems
chmod +x krems
```

#### Linux (AMD64)

```bash
curl -L https://github.com/mreider/krems/releases/download/vX.X.X/krems-linux-amd64 -o krems
chmod +x krems
```

#### Linux (ARM64)

```bash
curl -L https://github.com/mreider/krems/releases/download/vX.X.X/krems-linux-arm64 -o krems
chmod +x krems
```

#### Windows

Download `krems-windows-amd64.exe`. You can rename it to `krems.exe` if you like, and place it somewhere on your `PATH`.

### Put Krems in Your PATH

To run Krems from any directory, move the file into a folder that’s on your system `PATH`. On macOS/Linux, for example:

```bash
mv krems /usr/local/bin/
```
*(Adjust the directory if needed.)*

On Windows, you can place `krems.exe` in a directory like `C:\Program Files\Krems` and add that to your PATH environment variable.

---

## Basic Usage

Krems has three primary commands:

1. **`krems --init`**  
   Creates a starter directory structure with sample Markdown files, a `config.yaml`, and example images in a folder named `markdown/`. You’ll see a default site about “Mollusks” (which you can replace). This command embeds Bootstrap files as well, so your site is ready to go.

2. **`krems --build`**  
   Reads the `markdown/` directory and your `config.yaml`, then generates a static site in the `docs/` folder. That includes:  
   - Converting `.md` to `.html`  
   - Copying images, JS, and CSS  
   - Building an RSS feed
   - Creating a CNAME file for Github pages
   - Generating a `404.html`
   - Creating nice-looking Bootstrap-based pages with a responsive NAV bar

3. **`krems --run`**  
   Spawns a local HTTP server to serve the contents of `docs/` on `http://localhost:8080`. It also logs requests (200, 404, etc.). Use this for a quick local preview before publishing your site.

---

## Config File

When you run `krems --init`, you get a `config.yaml` with a structure like this:

```yaml
website:
  url: "example.com"
  name: "Mollusks"

menu:
  - title: "Home"
    path: "index.md"
  - title: "Articles"
    path: "articles/index.md"
```

- **`website.url`**: The base URL for your site (used in RSS feed links).  
- **`website.name`**: The display name of your site (used in the NAV bar).  
- **`menu`**: A list of links for your NAV bar. Each link has a `title` (the text that appears) and a `path` (the `.md` file in your `markdown/` folder, e.g. `index.md` or `articles/index.md`).

When you run `--build`, Krems locates each Markdown file referenced in the `menu` and rewrites it to a proper HTML link in the NAV.

---

## Markdown Files & Front Matter

### Directory Structure

By default, `krems --init` creates:

- `markdown/`  
  - `index.md`  
  - `about.md`  
  - `articles/`  
    - `index.md`  
    - `article1.md`  
- `config.yaml`

You can add more `.md` files anywhere under `markdown/`.  

### Front Matter

At the top of each Markdown file, you can optionally include a YAML block, for example:

```yaml
---
title: Another Octopus Poem
date: 2012-06-01
type: normal
description: An octopus poem with mention of crabs
image: images/mollusk.png
---
```

- **`title`** (required) – The display title of the page.  
- **`type`** (optional) – Either `normal` (default) or `list`.  
  - A `list` page automatically displays links to sibling pages in the same directory, grouped by date (descending).  
- **`date`** (optional) – If present, used for chronological listings and RSS pubDate. Format: `YYYY-MM-DD`.  
- **`description`** (optional) – Used in the `<meta name="description">` tag and for social previews.  
- **`image`** (optional) – Placed at the top of the generated page, plus used as the social preview image (OpenGraph/Twitter Card).  

**Below** that front matter is your normal Markdown body. Krems converts that to HTML and injects it into the page template.

---

## Customizing Your Site

1. **Edit `config.yaml`** – Update your site name, URL, and menu.  
2. **Edit or create Markdown pages** in `markdown/`. You can add images to `markdown/images/`.  
3. **Rebuild** with `krems --build`. Check the output in the `docs/` folder.  
4. **Locally test** with `krems --run`. Visit [http://localhost:8080](http://localhost:8080).  

When you’re satisfied, you can upload the `docs/` directory to GitHub Pages (or any other static hosting) and enjoy your new site.

---

## Other Things

- **RSS** – Krems generates an `rss.xml` that lists pages with valid `date` fields in descending order.  
- **404 Page** – A `404.html` is placed at the root of `docs/` so GitHub Pages (and other static hosts) can serve a custom “Not Found” message.  
- **Responsive** – Krems uses Bootstrap for layout and mobile responsiveness.  
- **Local Navigation** – The nav links are rewritten to use correct relative or absolute paths so your site works in subdirectories or at a domain root.

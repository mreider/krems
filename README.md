# Krems

[A little blog that explains Krems](https://mreider.com/tech/krems-static-blog-generator/).

**Krems** is a simple, lightweight static site generator written in Go. It converts Markdown files into responsive HTML pages, complete with navigation menus, optional blog listings, and an RSS feed. Think of it like Hugo or Jekyll, but stripped down to the essentials.

## Features

- Minimal configuration: a single `config.yaml` for site name, URL, and menu structure
- Friendly “front matter” in each Markdown file for title, description, etc.
- Automatic generation of “list” pages that group sub-pages by date
- Responsive output using Bootstrap (bundled automatically by `krems --build`)
- Integrated local server (`--run`) for quick previews
- 404 page generation for GitHub Pages or other static hosts
- RSS feed for dated pages
- Automatic CNAME file created for Github pages

---

## Simplified Usage with GitHub Actions (Recommended for New Users)

For users who prefer a hands-off approach without needing to install or run `krems` locally, a GitHub Actions workflow can automate the entire process of building and deploying your site.

**How it Works:**

1.  **Create Markdown Files:** You create your `.md` files and any static assets (like images in an `images/` folder, custom JavaScript in `js/`, etc.) directly in the root of your GitHub repository. Core CSS (like Bootstrap) is handled automatically by `krems`. Subfolders for organizing markdown content are also supported (e.g., `my-articles/article1.md`).
2.  **GitHub Action:** A provided GitHub workflow (see below) will:
    *   Automatically download the latest `krems` binary.
    *   If a `config.yaml` is missing, it generates one with default settings:
        *   Site Name: "My Blog"
        *   Website URL: Your default GitHub Pages URL (e.g., `https://your-username.github.io/your-repo-name/`)
        *   Menu: "Home" link, plus links for each top-level folder (e.g., a folder named `my-articles` becomes a "My Articles" menu item).
    *   Run `krems` to build your site into a temporary `/docs` folder (this folder won't be in your source branch).
    *   Deploy the contents of this temporary `/docs` folder to a special branch named `gh-pages`.
    *   If you later update `config.yaml` with a custom domain, the workflow will automatically create the `CNAME` file within the `gh-pages` branch.
3.  **Local Development & `.gitignore`:**
    *   Your main branch will only contain your source files (markdown, `config.yaml`, images, etc.). It will *not* contain the generated `/docs` folder.
    *   It's highly recommended to add `/docs/` to a `.gitignore` file in your project root. This prevents the locally generated build output from being accidentally committed to your source branch. Create a file named `.gitignore` in the root of your repository (if it doesn't exist) and add the following line:
        ```
        /docs/
        ```
4.  **GitHub Pages Configuration:**
    *   After the workflow runs for the first time, it will create the `gh-pages` branch.
    *   Go to your repository's "Settings" tab, then "Pages" in the sidebar.
    *   Under "Build and deployment", for "Source", select "Deploy from a branch".
    *   For "Branch", select `gh-pages` and keep the folder as `/ (root)`. Click "Save".
    *   Your site will then be served from the `gh-pages` branch.
5.  **Customization:**
    *   After the first workflow run, `config.yaml` will be in your main repository branch. You can edit it to change the site name, URL, or menu structure. The workflow will respect your changes.
    *   To use a custom domain, update the `website` URL in `config.yaml` and configure the custom domain in your repository's "Settings" -> "Pages" section. The workflow will handle the `CNAME` file on the `gh-pages` branch.

**Setting up the GitHub Workflow:**

1.  In your GitHub repository (on your main branch, e.g., `main`), create a file named `.github/workflows/deploy-krems-site.yml`.
2.  Paste the following content into it:

```yaml
name: Build and Deploy Krems Site

on:
  push:
    branches:
      - main # Or your default branch
  workflow_dispatch:

jobs:
  build-deploy:
    runs-on: ubuntu-latest
    permissions:
      contents: write
      pages: write
      id-token: write

    steps:
      - name: Checkout repository
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Install jq
        run: sudo apt-get update && sudo apt-get install -y jq

      - name: Download latest Krems binary
        run: |
          set -e
          KREMS_ASSET_NAME="krems-linux-amd64"
          LATEST_KREMS_URL=$(curl -s https://api.github.com/repos/mreider/krems/releases/latest | \
                             jq -r ".assets[] | select(.name | endswith(\"${KREMS_ASSET_NAME}\")) | .browser_download_url")
          if [ -z "$LATEST_KREMS_URL" ] || [ "$LATEST_KREMS_URL" == "null" ]; then
            echo "Error: Could not find ${KREMS_ASSET_NAME} in the latest release of mreider/krems."
            exit 1
          fi
          echo "Downloading Krems from: $LATEST_KREMS_URL"
          curl -sL -o krems "$LATEST_KREMS_URL"
          chmod +x ./krems
          echo "Krems version: $(./krems --version || echo 'version flag not supported')"

      - name: Configure Git User
        run: |
          git config user.name "GitHub Action Bot"
          git config user.email "actions@github.com"
          
      - name: Generate config.yaml if it does not exist
        id: generate_config
        run: |
          set -e
          if [ -f config.yaml ]; then
            echo "config.yaml found. Skipping generation."
          else
            echo "config.yaml not found. Generating initial version..."
            SITE_NAME="My Blog"
            REPO_OWNER=$(echo "${{ github.repository }}" | cut -d'/' -f1)
            REPO_NAME=$(echo "${{ github.repository }}" | cut -d'/' -f2)
            WEBSITE_URL="https://${REPO_OWNER}.github.io/${REPO_NAME}/"
            echo "name: \"${SITE_NAME}\"" > config.yaml
            echo "website: \"${WEBSITE_URL}\"" >> config.yaml
            echo "menu:" >> config.yaml
            echo "  - name: Home" >> config.yaml
            echo "    url: /" >> config.yaml
            find . -maxdepth 1 -mindepth 1 -type d \
              ! -name ".*" ! -name "docs" ! -name "node_modules" ! -name "vendor" \
              -print0 | while IFS= read -r -d $'\0' dir; do
              DIR_NAME=$(basename "$dir")
              MENU_NAME=$(echo "$DIR_NAME" | sed -e 's/[_-]/ /g' -e 's/\b\(.\)/\u\1/g')
              echo "  - name: \"${MENU_NAME}\"" >> config.yaml
              echo "    url: /${DIR_NAME}/" >> config.yaml
            done
            echo "Generated config.yaml:"
            cat config.yaml
            git add config.yaml
            if git diff --staged --quiet; then
              echo "No changes to commit for config.yaml."
            else
              git commit -m "feat: Generate initial config.yaml [skip ci]"
              git push origin HEAD:${{ github.ref_name }}
            fi
          fi

      - name: Run Krems to build website
        run: |
          set -e
          ./krems --build # Ensure --build is used if main.go expects a command

      - name: Generate CNAME file for custom domain (if applicable)
        run: |
          set -e
          WEBSITE_LINE=$(grep '^website:' config.yaml || echo "")
          if [ -z "$WEBSITE_LINE" ]; then echo "Warning: 'website:' not in config.yaml."; exit 0; fi
          WEBSITE_URL_FROM_CONFIG=$(echo "$WEBSITE_LINE" | sed -n 's/website: "\(.*\)"/\1/p')
          if [ -z "$WEBSITE_URL_FROM_CONFIG" ]; then WEBSITE_URL_FROM_CONFIG=$(echo "$WEBSITE_LINE" | sed -n "s/website: '\(.*\)'/\1/p"); fi
          if [ -z "$WEBSITE_URL_FROM_CONFIG" ]; then WEBSITE_URL_FROM_CONFIG=$(echo "$WEBSITE_LINE" | sed -n 's/website: \(.*\)/\1/p'); fi
          if [ -z "$WEBSITE_URL_FROM_CONFIG" ]; then echo "Warning: Could not parse 'website' URL."; exit 0; fi
          if [[ "$WEBSITE_URL_FROM_CONFIG" != *"github.io"* ]] && [[ "$WEBSITE_URL_FROM_CONFIG" == "http"* ]]; then
            CUSTOM_DOMAIN=$(echo "$WEBSITE_URL_FROM_CONFIG" | sed -e 's|^https\?://||' -e 's|/$||' -e 's|/.*$||')
            if [ -n "$CUSTOM_DOMAIN" ]; then echo "$CUSTOM_DOMAIN" > docs/CNAME; echo "CNAME created: $CUSTOM_DOMAIN"; fi
          else
            rm -f docs/CNAME
          fi

      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v4
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: ./docs
          user_name: 'github-actions[bot]'
          user_email: 'github-actions[bot]@users.noreply.github.com'
          commit_message: 'Deploy: ${{ github.event.head_commit.message }} [skip ci]'
          publish_branch: gh-pages # Explicitly deploy to gh-pages branch
```

With this setup:
1. Add `/docs/` to your `.gitignore` file.
2. Configure GitHub Pages to serve from the `gh-pages` branch (from the `/ (root)` of that branch).
3. Push your markdown changes to your main branch, and your site updates automatically on the `gh-pages` branch!

---

## For Krems Developers/Contributors: Sample Site at Root

Please note that the root of the `krems` repository itself now serves as the source material for the sample website (e.g., for `https://mreider.github.io/krems/`). This sample site is automatically built and deployed to the `gh-pages` branch by the `.github/workflows/deploy-krems-site.yml` workflow.

-   The `config.yaml` file located at the root of this repository is used for configuring this sample site.
-   The main project `README.md` and other Go project files (`.go`, `go.mod`, etc.) are ignored by `krems` when it builds the sample site from the root.

---

## Manual Installation & Usage (Advanced)

If you prefer to run `krems` locally or need more control:

### Installation

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

Krems has two primary commands:

1.  **`krems --build`**  
    Reads markdown files and assets (like `images/`, `js/`) from the current directory (and its subdirectories, excluding `docs/`, `.git/`, etc.) and your `config.yaml` (if present), then generates a static site in the `docs/` folder. This includes:
    - Converting `.md` to `.html`
    - Automatically generating necessary CSS (like Bootstrap) into `docs/css/`.
    - Copying assets from `images/` and `js/` (if they exist at the root) into `docs/`.
    - Building an RSS feed
    - Creating a CNAME file for GitHub Pages (if a custom domain is in `config.yaml`)
    - Generating a `404.html`
    - Creating nice-looking Bootstrap-based pages with a responsive NAV bar

2.  **`krems --run`**  
    Spawns a local HTTP server to serve the contents of `docs/` on `http://localhost:8080`. It also logs requests (200, 404, etc.). Use this for a quick local preview *after* running `krems --build`.

---

## Config File

A `config.yaml` file at the root of your project is used to configure your site's name, URL, and menu structure. If you don't have one and are using the GitHub Actions workflow described above, a basic one will be generated for you.

A typical `config.yaml` looks like this:

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
- **`menu`**: A list of links for your NAV bar. Each link has a `title` (the text that appears) and a `path` (the `.md` file relative to your project root, e.g. `index.md` or `articles/index.md`).

When you run `--build`, Krems locates each Markdown file referenced in the `menu` and rewrites it to a proper HTML link in the NAV.

---

## Markdown Files & Front Matter

### Directory Structure (Post v0.2.0)

`krems` now expects markdown files and asset folders (like `images/`, `js/`) to be at the root of your project.

A typical structure might look like:

- `index.md` (your homepage)
- `about.md`
- `articles/` (a folder for your articles)
  - `index.md` (optional, could be a list page for articles)
  - `my-first-article.md`
  - `another-topic.md`
- `images/`
  - `photo.jpg`
- `js/`
  - `custom.js` (if you have custom JavaScript)
- `config.yaml`

Core CSS (Bootstrap) is automatically included by `krems` during the build process; you do not need to create or manage a `css/` folder at the root for these default styles. If you wish to add your *own* custom CSS, you would typically create a `custom.css` file (e.g., in a root `css/` folder you create yourself) and link to it from your HTML templates (an advanced customization).

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
2. **Edit or create Markdown pages** in your project root or subfolders (e.g., `articles/`). You can add images to an `images/` folder at the root.
3. **Rebuild** with `krems --build`. Check the output in the `docs/` folder.
4. **Locally test** with `krems --run`. Visit [http://localhost:8080](http://localhost:8080).

When you’re satisfied, you can upload the `docs/` directory to GitHub Pages (or any other static hosting) or rely on the GitHub Action to do it for you.

---

## Breaking Changes (from versions prior to v0.2.0)

-   **Markdown and Asset Location:** `krems` (from v0.2.0 or the version incorporating these simplified workflow changes) no longer looks for content within a `markdown/` subdirectory.
    *   All your markdown files (`.md`) should be in the root of your project or in subdirectories directly under the root (e.g., `my-articles/article.md`).
    *   User-provided static assets (like `js/`, `images/` folders) should also be at the root level. Core CSS is handled internally.
    *   If you run `krems` locally and an old `markdown/` directory is detected, `krems` will display a warning message guiding you to move your content. The build will proceed by looking for content at the root, ignoring the `markdown/` directory.
-   **`krems --init` Command Removed:** The `--init` command has been removed. Users should now create their markdown files and optionally a `config.yaml` directly. The GitHub Actions workflow (`deploy-krems-site.yml`) can generate a default `config.yaml` if one is not present in a user's repository.

---

## Other Things

- **RSS** – Krems generates an `rss.xml` that lists pages with valid `date` fields in descending order.  
- **404 Page** – A `404.html` is placed at the root of `docs/` so GitHub Pages (and other static hosts) can serve a custom “Not Found” message.  
- **Responsive** – Krems uses Bootstrap for layout and mobile responsiveness.  
- **Local Navigation** – The nav links are rewritten to use correct relative or absolute paths so your site works in subdirectories or at a domain root.

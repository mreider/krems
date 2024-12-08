## Krems Static Site Generator

### Overview  
Krems is a simple Ruby-based static site generator designed to convert Markdown files into HTML. The generated files and assets are placed in a `published` directory, making it easy to preview locally during development and compatible with GitHub Pages for deployment.

I love Hugo, Hexo, Jekyll, Astro, and other static site generators, but I wanted something much simpler—something that would support a classless stylesheet, work with standard Markdown export libraries like Redcarpet, and avoid any magic. Krems gives you total control without unnecessary complexity.

For an example of a complete site built with Krems, check out the repository [https://github.com/mreider/mreider.com](https://github.com/mreider/mreider.com) or visit the live site at [mreider.com](https://mreider.com)

---

### Getting Started

1. **Download and Setup:**

 **Prerequisite:** Ensure you have Ruby installed on your system.
 
   - Download the latest release from the [GitHub Releases page](https://github.com/mreider/krems/releases).
   - Unzip the release to a directory of your choice.
   - Open a terminal and `cd` into the directory.


2. **Install Dependencies:**
   ```bash
   bundle install
   ```

3. **Initialize a New Project:**
   ```bash
   ruby krems.rb --init
   ```
   This creates all necessary files and directories for a simple example site.

4. **Serve the Site:**
   ```bash
   ruby krems.rb --serve
   ```
   This command builds the site and starts a local development server. Visit `http://localhost:4567` in your browser to see the site.

5. **Explore the Generated Files:**
   - Look around the `markdown/`, `css/`, and other directories to see how the example site was created.
   - Check out the `index.md` file in the `markdown/` directory, which includes a post listing example.

---

### Example Workflow for GitHub Pages

To deploy your site to GitHub Pages, you can use the following workflow file. This file assumes your `published/` directory contains the generated HTML.

#### `.github/workflows/pages.yml`

```yaml
name: Build and Deploy Krems Site to GitHub Pages

on:
  push:
    branches:
      - main 
  workflow_dispatch:

permissions:
  contents: read
  pages: write
  id-token: write

concurrency:
  group: "pages"
  cancel-in-progress: true

defaults:
  run:
    shell: bash

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Repository
        uses: actions/checkout@v3
        with:
          submodules: recursive

      - name: Setup Pages
        id: pages
        uses: actions/configure-pages@v2

      - name: Setup Ruby
        uses: ruby/setup-ruby@v1
        with:
          ruby-version: 3.1

      - name: Install Dependencies
        run: bundle install

      - name: Build Krems Site
        env:
          BASE_URL: ${{ steps.pages.outputs.base_url }}
        run: ruby krems.rb --build

      - name: Upload Krems Artifact
        uses: actions/upload-pages-artifact@v1
        with:
          path: ./published

  deploy:
    environment:
      name: github-pages
      url: ${{ steps.deployment.outputs.page_url }}
    runs-on: ubuntu-latest
    needs: build
    steps:
      - name: Deploy to GitHub Pages
        id: deployment
        uses: actions/deploy-pages@v1
```
---

### Directory Structure

When you initialize a new project, Krems generates the following structure:

```
/
├── markdown/        # Directory for Markdown files
│   ├── index.md     # Main Markdown file (default entry point)
│   ├── example/     # Example directory with two posts
│       ├── post1.md # Example post 1
│       └── post2.md # Example post 2
├── css/             # Directory for CSS files (static assets)
│   └── styles.css   # Default stylesheet
├── images/          # Directory for images (static assets)
├── published/       # Directory where generated HTML and assets are published
├── krems.rb         # Main Ruby script
├── defaults.toml    # Optional defaults for front matter (TOML format)
├── config.toml      # Global configuration for base URL and CSS file
└── Gemfile          # Gem dependencies for the project
```

---

### Features

1. **Markdown to HTML Conversion**
   - Converts Markdown files to HTML using the Redcarpet library.
   - Extracts front matter written in TOML format for metadata such as `title`, `author`, `date`, `summary`, etc.

2. **Front Matter Handling**
   - Supports TOML front matter at the top of Markdown files.
   - Merges defaults from `defaults.toml` with each file’s front matter.

3. **Global Configuration**
   - The `config.toml` file allows you to set:
     - `url`: Base URL for the site.
     - `css`: Specify the CSS file to be used for styling.

4. **Post List Generation**
   - Generates a list of posts within a folder using the `{{ list_posts(folder_name) }}` placeholder.

5. **Menu Generation**
   - Parses a `menu` field from `defaults.toml` or front matter to generate navigation menus.

6. **Static Assets**
   - Automatically links CSS and image files in the generated HTML `<head>` section.

7. **Live Server**
   - `--serve` builds the site and starts a local server.
   - Automatically rebuilds the site on file changes using the `listen` gem.

---

### Support and Questions

If you have any questions, feel free to open an issue in the [GitHub Issues](https://github.com/mreider/krems/issues) section of the repository.

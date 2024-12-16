## Krems Static Site Generator

### Overview

Krems is a lightweight, Ruby-based static site generator that converts Markdown files into responsive HTML. Designed for simplicity, Krems combines classless and Bootstrap styling options with standard Markdown export libraries like Redcarpet, giving you complete control over your site's structure and functionality without unnecessary complexity or magic.

Visit the [repository](https://github.com/mreider/mreider.com) for an example of a site built with Krems or check out the live site at [mreider.com](https://mreider.com).

---

### Getting Started

#### Prerequisites

- Ruby (3.1 or higher recommended)
- Bundler gem (`gem install bundler`)

#### Setup and Installation

1. **Clone the Repository**  
   Clone or download the Krems repository:
   ```bash
   git clone https://github.com/mreider/krems.git
   cd krems
   ```

2. **Install Dependencies**  
   Install the required gems:
   ```bash
   bundle install
   ```

3. **Initialize a New Project**  
   Create a new Krems project in the current directory:
   ```bash
   ruby krems.rb --init
   ```
   This generates all required directories and files for a basic site, including example Markdown files, default configurations, and a minimal stylesheet.

4. **Preview Locally**  
   Build the site and start a local development server:
   ```bash
   ruby krems.rb --serve
   ```
   Access your site at `http://127.0.0.1:4567`. The server automatically rebuilds the site whenever you make changes to Markdown, CSS, or image files.

5. **Build for Production**  
   Generate the static site for deployment:
   ```bash
   ruby krems.rb --build
   ```
   The output will be in the `published/` directory.

---

### Features

1. **Markdown to HTML Conversion**  
   Converts Markdown files to HTML using the Redcarpet library. Supports:
   - Tables
   - Autolinks
   - Fenced code blocks
   - TOML front matter for metadata

2. **TOML Front Matter**  
   Add metadata such as `title`, `author`, `date`, `summary`, and `image` directly in Markdown files. Default metadata can be defined in `defaults.toml`.

3. **Dynamic Menu Generation**  
   Automatically generate navigation menus using the `menu` field in `defaults.toml` or individual file front matter. Supports nested dropdown menus.

4. **Post Listing by Year**  
   Generate a chronological list of posts in any folder using the `{{ list_posts(folder_name) }}` placeholder.

5. **Bootstrap Styling**  
   Built-in Bootstrap integration for responsive layouts, including navigation menus, post metadata, and overall styling.

6. **Live Server with Auto-Rebuild**  
   The `--serve` command starts a local server and rebuilds your site automatically whenever changes are detected in Markdown, CSS, or images.

7. **Static Asset Handling**  
   Automatically copies CSS, images, and other static files to the `published/` directory.

8. **Open Graph Meta Tags**  
   Automatically generates Open Graph meta tags from front matter for improved social media sharing.

Other things it does:

- **Custom Link Conversion**: Converts `.md` links in Markdown files to `.html` links in the generated site.
- **Nested Menus**: Supports multi-level navigation menus using dropdowns.
- **Flexible Configuration**: Define the base URL, CSS file, and other settings in `config.toml`.

---

### Directory Structure

When you initialize a Krems project, the following structure is created:

```
/
├── markdown/        # Markdown source files
│   ├── index.md     # Main entry point
│   ├── example/     # Example posts
│       ├── post1.md # Example post 1
│       └── post2.md # Example post 2
├── css/             # Custom CSS files
│   └── styles.css   # Default stylesheet
├── images/          # Image assets
├── published/       # Generated static files
├── krems.rb         # Main Ruby script
├── defaults.toml    # Default front matter
├── config.toml      # Global configuration
└── Gemfile          # Gem dependencies
```

---

### Example Workflow for GitHub Pages

To deploy your site to GitHub Pages, use the following GitHub Actions workflow:

#### `.github/workflows/pages.yml`

```yaml
name: Deploy Krems Site to GitHub Pages

on:
  push:
    branches:
      - main
  workflow_dispatch:

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@v3

      - name: Setup Ruby
        uses: ruby/setup-ruby@v1
        with:
          ruby-version: 3.1

      - name: Install Dependencies
        run: bundle install

      - name: Build Krems Site
        run: ruby krems.rb --build

      - name: Deploy to GitHub Pages
        uses: actions/upload-pages-artifact@v1
        with:
          path: ./published
```

---

### Support and Contributions

Have a question or found a bug? Open an issue on the [GitHub Issues](https://github.com/mreider/krems/issues) page.

Pull requests are welcome for improvements or new features.
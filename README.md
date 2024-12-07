# Krems Static Site Generator

## Overview

Krems is a simple Ruby-based static site generator designed to convert Markdown files into HTML. The generated files and assets are placed in a `published` directory, making it easy to preview locally during development and compatible with GitHub Pages for deployment.

---

## Directory Structure

The generator expects a specific directory structure for inputs and outputs:

```
/
├── markdown/        # Directory for Markdown files
│   ├── index.md     # Main Markdown file (default entry point)
│   ├── folder_name/ # Subfolders containing additional Markdown files
│       └── file.md  # Markdown files inside subfolders
├── css/             # Directory for CSS files (static assets)
│   └── styles.css   # Default stylesheet
├── images/          # Directory for images (static assets)
│   ├── favicon.ico  # Favicon file
│   ├── *.png        # Other image assets
├── published/       # Directory where generated HTML and assets are published
├── krems.rb         # Main Ruby script
├── defaults.toml    # Optional defaults for front matter (TOML format)
├── Gemfile          # Gem dependencies for the project
└── .github/         # GitHub Actions workflows
    └── workflows/
        └── pages.yml # GitHub Pages deployment workflow
```

---

## Features

### 1. **Markdown to HTML Conversion**
- Converts Markdown files to HTML using the [Redcarpet](https://github.com/vmg/redcarpet) library.
- Extracts front matter written in TOML format for metadata such as title, author, date, summary, etc.

### 2. **Front Matter Handling**
- Supports TOML front matter at the top of Markdown files.
- Merges defaults from `defaults.toml` with each file’s front matter.

Example:
```markdown
+++
title = "Example Post"
author = "John Doe"
date = "2024-12-06"
summary = "This is an example summary."
+++

# Example Post
This is the content of the post.
```

### 3. **Static Assets**
- Automatically links CSS and image files in the generated HTML `<head>` section.
- Includes default meta tags for Open Graph (`og:title`, `og:description`, etc.) if specified in front matter.

### 4. **Post List Generation**
- Automatically generates lists of posts within a folder using the `{{ list_posts(folder_name) }}` placeholder in Markdown.
- Links are formatted as absolute paths to ensure compatibility with GitHub Pages.

Example Markdown:
```markdown
{{ list_posts(tech) }}
```

Generated HTML:
```html
<h4>2024</h4>
<ul>
  <li><a href="/tech/article1.html">Article1</a></li>
  <li><a href="/tech/article2.html">Article2</a></li>
</ul>
```

### 5. **Menu Generation**
- Parses a `menu` field from front matter to generate navigation menus.

Example Front Matter:
```toml
menu = [
  { path = "/about.md", name = "About" },
  { path = "/contact.md", name = "Contact" }
]
```

Generated Menu:
```html
<table>
  <tr>
    <td><h4><a href="/about.html">About</a></h4></td>
    <td><h4>•</h4></td>
    <td><h4><a href="/contact.html">Contact</a></h4></td>
  </tr>
</table>
```

---

## Default Behavior

- Creates a default `index.md` file in the `markdown` folder if it doesn’t exist.
- Outputs all generated HTML files, assets, and directories to the `published` directory for local development and deployment.
- Deletes any files in the `published` directory before generating new content.

---

## Running the Generator

1. Install dependencies:
   ```bash
   bundle install
   ```

2. Generate the site and start the server:
   ```bash
   ruby krems.rb
   ```

3. Access the generated site locally:
   ```
   http://localhost:4567
   ```

4. All static files are located in the `published` directory. This can be used for deployment.

---

## Deployment to GitHub Pages

### Using GitHub Actions
The repository includes a GitHub Actions workflow (`.github/workflows/pages.yml`) for deploying the `published` directory to GitHub Pages.

#### Workflow Steps:
1. **Build the Site**: The workflow runs `krems.rb` to generate the site in the `published` directory.
2. **Deploy to GitHub Pages**: The `published` directory is deployed to the `gh-pages` branch.

### Setting Up GitHub Pages
1. Go to your repository **Settings**.
2. Navigate to the **Pages** section.
3. Set the source to `gh-pages` under the **Branch** dropdown.
4. Your site will be available at `https://<your-username>.github.io/<repository-name>`.

---

## Configuration

- **defaults.toml**: Specifies global defaults for front matter values.
- **CSS and Images**: Place custom stylesheets in `css/` and image files in `images/`.

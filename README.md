## Krems Static Site Generator

**Overview**  
Krems is a simple Ruby-based static site generator designed to convert Markdown files into HTML. The generated files and assets are placed in a `published` directory, making it easy to preview locally during development and compatible with GitHub Pages for deployment.

---

## Directory Structure

The generator expects the following directory structure:

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
├── config.toml      # Global configuration for base URL and CSS file
├── Gemfile          # Gem dependencies for the project
└── .github/         # GitHub Actions workflows
    └── workflows/
        └── pages.yml # GitHub Pages deployment workflow
```

---

## Features

1. **Markdown to HTML Conversion**
   - Converts Markdown files to HTML using the Redcarpet library.
   - Extracts front matter written in TOML format for metadata such as `title`, `author`, `date`, `summary`, etc.

2. **Front Matter Handling**
   - Supports TOML front matter at the top of Markdown files.
   - Merges defaults from `defaults.toml` with each file’s front matter.

   Example:

   ```toml
   +++
   title = "Example Post"
   author = "John Doe"
   date = "2024-12-06"
   summary = "This is an example summary."
   +++
   ```

   Content below the front matter will be parsed as Markdown.

3. **Global Configuration**
   - The `config.toml` file allows users to set:
     - `url`: Base URL for the site (e.g., for deployment on GitHub Pages or local preview).
     - `css`: Specify the CSS file to be used for styling.

   Example:

   ```toml
   url = "https://example.com/"
   css = "custom.css"
   ```

4. **Static Assets**
   - Automatically links CSS and image files in the generated HTML `<head>` section.
   - Includes default meta tags for Open Graph (e.g., `og:title`, `og:description`) if specified in the front matter.

5. **Post List Generation**
   - Generates a list of posts within a folder using the `{{ list_posts(folder_name) }}` placeholder in Markdown.

   Example:

   Markdown:
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

6. **Menu Generation**
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

7. **Live Server**
   - Runs a local server with the `--serve` flag.
   - Automatically rebuilds the site on file changes using the `listen` gem.

---

## Running the Generator

1. **Install Dependencies:**
   ```bash
   bundle install
   ```

2. **Generate the Site:**
   ```bash
   ruby krems.rb --build
   ```

3. **Run a Local Server:**
   ```bash
   ruby krems.rb --serve
   ```

4. **Access the Site:**
   ```
   http://localhost:4567
   ```

---

## Deployment to GitHub Pages

### Using GitHub Actions

1. The repository includes a workflow file (`.github/workflows/pages.yml`) for deploying the `published` directory to GitHub Pages.
2. Configure GitHub Pages to use the `gh-pages` branch.

---

## Configuration

- **`defaults.toml`:** Specifies global defaults for front matter values.
- **`config.toml`:** Specifies global settings like `url` and `css`.
- **CSS and Images:** Place custom stylesheets in the `css/` directory and image files in the `images/` directory.
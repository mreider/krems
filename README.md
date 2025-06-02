# Krems: Simple Static Sites for Markdown Blogs

## 1. What is Krems?

Krems is a straightforward static site generator designed for creating clean, fast markdown-based blogs and websites.
- **Simple by Design**: Focus on your content; Krems handles the rest.
- **Markdown-Powered**: Write your posts and pages in familiar Markdown.
- **Fixed Styling**: Comes with a clean, responsive design (Bootstrap CSS) included, so you don't have to worry about CSS.
- **All-Inclusive**: Generates HTML, navigation, RSS feeds, and handles assets.
- **Obsidian Friendly**: An Obsidian plugin is planned to make content creation even smoother.

## 2. How It Works: Get Started with the Example Site

The easiest way to understand Krems is to explore the example site.

1.  **Clone the Example Site:**
    ```bash
    git clone https://github.com/mreider/krems-example.git
    cd krems-example
    ```
2.  **Explore the Structure:**
    Look at the `config.yaml` file, the markdown files (`.md`), and how content is organized. This will give you a feel for how Krems projects are structured.

## 3. Build It Locally

To build the site from the markdown files into static HTML:

- If you have Krems installed globally:
  ```bash
  krems --build
  ```
- If you are running Krems from its source code (e.g., developing Krems itself):
  ```bash
  go run . --build
  ```
This command will generate the website into a `docs/` folder.

## 4. Run It Locally

To preview your site locally after building it:

- If you have Krems installed globally:
  ```bash
  krems --run
  ```
- If you are running Krems from its source code:
  ```bash
  go run . --run
  ```
This will start a local web server, typically at `http://localhost:8080`. The `krems --run` command automatically rebuilds the site using settings suitable for local development (like `devPath: "/"` in your `config.yaml`).

## 5. Deploy Your Own Site with GitHub Pages

You can easily host your Krems site for free using GitHub Pages. The `krems-example` repository includes a GitHub Actions workflow that automates building and deploying your site.

1.  **Create Your Repository on GitHub:**
    Go to GitHub and create a new, empty repository (e.g., `yourusername/my-new-blog`).

2.  **Clone the Example and Set Your Remote:**
    ```bash
    # Clone the example site into a temporary directory
    git clone https://github.com/mreider/krems-example.git my-blog-temp
    cd my-blog-temp

    # Change the remote URL to point to *your* new GitHub repository
    git remote set-url origin https://github.com/yourusername/my-new-blog.git # Replace with your repo URL

    # Push the example site's code to your new repository
    # (Use 'main' or your repository's default branch name)
    git push -u origin main
    ```

3.  **GitHub Actions Workflow:**
    The `krems-example` site includes a GitHub Actions workflow file (in `.github/workflows/`). When you push code to your repository's main branch, this workflow will:
    *   Automatically build your Krems site.
    *   Deploy the contents of the `docs/` folder to a special `gh-pages` branch in your repository.

4.  **Configure GitHub Pages:**
    *   In your GitHub repository, go to "Settings" > "Pages".
    *   Under "Build and deployment", for "Source", select "Deploy from a branch".
    *   For "Branch", choose `gh-pages` and ensure the folder is set to `/ (root)`. Click "Save".
    *   Your site will typically be available at `https://yourusername.github.io/my-new-blog/`.

## 6. Configure Your Site

After setting up your repository, you'll need to customize the configuration:

1.  **Edit `config.yaml`:**
    Open the `config.yaml` file in your repository.
    *   **`website.url`**: Change this to your site's URL.
        *   For a GitHub Pages site like `https://yourusername.github.io/my-new-blog/`, set it to this URL.
        *   If using a custom domain (e.g., `https://www.myawesomeblog.com`), set it to your custom domain.
    *   **`website.name`**: Set your desired site name (appears in the navigation bar).
    *   **`website.basePath`**: This is important for links to work correctly.
        *   If your site is at `https://yourusername.github.io/my-new-blog/`, set `basePath: "/my-new-blog/"`.
        *   If your repository is named `yourusername.github.io` and you want the site at `https://yourusername.github.io/`, set `basePath: "/"`
        *   If using a custom domain pointing to the root, set `basePath: "/"`
    *   **`website.devPath`**: For local development with `krems --run` or `krems --build` (locally), you'll typically want `devPath: "/"`. This is usually pre-configured in the example.
    *   Update the `menu` section to reflect your site's structure.

2.  **Custom Domain (CNAME):**
    If you want to use a custom domain (e.g., `www.myawesomeblog.com`):
    *   First, configure your custom domain in your repository's "Settings" > "Pages" section on GitHub.
    *   Then, simply set the `website.url` in your `config.yaml` to your full custom domain URL (e.g., `https://www.myawesomeblog.com`). Krems will automatically generate the necessary `CNAME` file during the build process.

That's it! You now have a simple, markdown-powered blog or website running with Krems.

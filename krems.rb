require 'fileutils'
require 'redcarpet'
require 'pathname'
require 'toml-rb'
require 'date'
require 'titleize'
require 'listen'
require 'optparse'
require 'rss'

MARKDOWN_DIR = "markdown"
CSS_DIR = "css"
IMAGES_DIR = "images"
PUBLISHED_DIR = "published"
CONFIG_FILE = "config.toml"

def normalize_url(url)
  url = url.to_s.strip
  url.end_with?("/") ? url : "#{url}/"
end

def initialize_project
  unless File.exist?(File.join(MARKDOWN_DIR, "index.md"))
    puts "Initializing new Krems project..."

    [MARKDOWN_DIR, CSS_DIR, IMAGES_DIR, PUBLISHED_DIR].each do |dir|
      FileUtils.mkdir_p(dir)
      puts "Created directory: #{dir}"
    end

    config_content = <<~TOML
      url = "http://127.0.0.1:4567/"
    TOML
    File.write(CONFIG_FILE, config_content)
    puts "Created file: #{CONFIG_FILE}"

    defaults_content = <<~TOML
      title = "Default Title"
      author = "Default Author"
      summary = "This is a default summary."
      menu = [
        { path = "/index.md", name = "Home" },
        { path = "/example/post1.md", name = "First Post" }
      ]
    TOML

    File.write("defaults.toml", defaults_content)
    puts "Created file: defaults.toml"

    index_content = <<~MARKDOWN
      +++
      title = "Welcome to Krems"
      +++

      This is the default index page. Below is a list of posts in the 'example' directory:

      {{ list_posts(example) }}
    MARKDOWN
    File.write(File.join(MARKDOWN_DIR, "index.md"), index_content)
    puts "Created file: markdown/index.md"

    example_dir = File.join(MARKDOWN_DIR, "example")
    FileUtils.mkdir_p(example_dir)
    puts "Created directory: markdown/example"

    example_post_1 = <<~MARKDOWN
      +++
      title = "First Example Post"
      author = "Krems"
      date = "#{Time.now.strftime('%Y-%m-%d')}"
      summary = "This is the first example post."
      +++

      This is the content of the first example post.
    MARKDOWN
    File.write(File.join(example_dir, "post1.md"), example_post_1)
    puts "Created file: markdown/example/post1.md"

    example_post_2 = <<~MARKDOWN
      +++
      title = "Second Example Post"
      author = "Krems"
      date = "#{(Time.now - 86400).strftime('%Y-%m-%d')}"
      summary = "This is the second example post."
      +++

      This is the content of the second example post.
    MARKDOWN
    File.write(File.join(example_dir, "post2.md"), example_post_2)
    puts "Created file: markdown/example/post2.md"
    puts "Krems project initialized successfully."
  end
end

def load_base_url(local = false)
  if local
    "http://127.0.0.1:4567/"
  elsif File.exist?(CONFIG_FILE)
    normalize_url(TomlRB.load_file(CONFIG_FILE)['url'] || "/")
  else
    "/"
  end
end

def clean_published_directory
  FileUtils.rm_rf(PUBLISHED_DIR)
  FileUtils.mkdir_p(PUBLISHED_DIR)
end


def load_defaults
  defaults_file = "defaults.toml"
  if File.exist?(defaults_file)
    defaults = TomlRB.load_file(defaults_file)
    defaults
  else
    {}
  end
end

def merge_defaults(front_matter, defaults)
  merged = front_matter.dup
  defaults.each do |key, value|
    if key == "menu" && front_matter[key].is_a?(Array) && value.is_a?(Array)
      merged[key] = (front_matter[key] + value).uniq
    else
      merged[key] = value unless merged.key?(key)
    end
  end
  merged
end

def parse_front_matter(content, defaults)
  if content.strip.start_with?("+++")
    front_matter, body = content.split("+++", 3)[1..2]
    merged_front_matter = merge_defaults(TomlRB.parse(front_matter), defaults)
    [merged_front_matter, body.strip]
  else
    [defaults, content.strip]
  end
end


def absolute_path(base_url, relative_path)
  base_url = normalize_url(base_url)
  relative_path = relative_path.sub(%r{^/}, "")
  "#{base_url}#{relative_path}"
end

def convert_links_to_html(content, base_url)
  content.gsub(/href="([^"]+)\.md(\?[^"]*|#[^"]*|)"/) do
    link_path = absolute_path(base_url, $1) + ".html" + $2
    "href=\"#{link_path}\""
  end
end

def update_image_links(content, base_url)
  content.gsub(/!\[([^\]]*)\]\((\/?images\/[^\)]+)\)/) do
    alt_text, image_path = $1, $2
    "![#{alt_text}](#{absolute_path(base_url, image_path)})"
  end
end

def generate_static_asset_links(base_url)
  <<~HTML
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="#{absolute_path(base_url, 'css/styles.css')}">
    <link href="https://fonts.googleapis.com/css2?family=Lato:wght@400;700&family=Raleway:wght@400;700&display=swap" rel="stylesheet">
    <link rel="icon" href="#{absolute_path(base_url, 'images/favicon.ico')}">
  HTML
end

def generate_header(base_url, front_matter)
  menu_items = front_matter['menu'] || []
  nav_links = menu_items.map do |entry|
    formatted_path = entry['path'].sub(/\.md$/, '.html')
    "<li class='nav-item'><a class='nav-link' href='#{absolute_path(base_url, formatted_path)}'>#{entry['name']}</a></li>"
  end.join

  quacker_enabled = File.exist?(CONFIG_FILE) && TomlRB.load_file(CONFIG_FILE)['quacker'] == true
  subscription_form = ""
  if quacker_enabled
    subscription_form = <<~HTML
      <form id="subscription-form" class="d-flex ms-auto">
        <input type="email" name="email" required placeholder="Enter your email" class="form-control me-2">
        <button type="submit" class="btn btn-primary">Subscribe</button>
      </form>
      <p id="status-message" class="ms-auto mt-2 text-success" style="display: none;"></p>
      <script>
        const form = document.getElementById("subscription-form");
        const statusMessage = document.getElementById("status-message");
        const checkSubscriptionStatus = async (email) => {
            const url = "https://quacker.eu/subscribe/1?email=" + encodeURIComponent(email);
            try {
                const response = await fetch(url);
                if (response.ok) {
                    const data = await response.json();
                    return data.is_subscribed;
                }
            } catch (error) {
                console.error("Error checking subscription status:", error);
            }
            return false;
        };
        form.addEventListener("submit", async (e) => {
            e.preventDefault();
            const emailInput = form.querySelector('input[name="email"]');
            const button = form.querySelector("button");
            const email = emailInput.value;
            button.textContent = "Please wait...";
            button.disabled = true;
            emailInput.disabled = true;
            try {
                const isSubscribed = await checkSubscriptionStatus(email);
                if (isSubscribed) {
                    form.style.display = "none";
                    statusMessage.style.display = "block";
                    statusMessage.textContent = "Already subscribed!";
                    return;
                }
                const url = "https://quacker.eu/subscribe/1";
                const response = await fetch(url, {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({ email })
                });
                if (response.ok) {
                    form.style.display = "none";
                    statusMessage.style.display = "block";
                    statusMessage.textContent = "Subscribed successfully!";
                } else {
                    form.style.display = "none";
                    statusMessage.style.display = "block";
                    statusMessage.textContent = "Something went wrong. Please try again later.";
                }
            } catch (error) {
                console.error("Error submitting form:", error);
                form.style.display = "none";
                statusMessage.style.display = "block";
                statusMessage.textContent = "Something went wrong. Please try again later.";
            }
        });
      </script>
    HTML
  end

  <<~HTML
    <nav class="navbar navbar-expand-lg navbar-light bg-light">
      <div class="container-fluid">
        <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#navbarNav" aria-controls="navbarNav" aria-expanded="false" aria-label="Toggle navigation">
          <span class="navbar-toggler-icon"></span>
        </button>
        <div class="collapse navbar-collapse" id="navbarNav">
          <ul class="navbar-nav">
            #{nav_links}
          </ul>
          #{subscription_form}
        </div>
      </div>
    </nav>
  HTML
end

def copy_static_assets
  FileUtils.mkdir_p(File.join(PUBLISHED_DIR, CSS_DIR))
  FileUtils.mkdir_p(File.join(PUBLISHED_DIR, IMAGES_DIR))
  FileUtils.cp_r(Dir.glob(File.join(CSS_DIR, "*")), File.join(PUBLISHED_DIR, CSS_DIR))
  FileUtils.cp_r(Dir.glob(File.join(IMAGES_DIR, "*")), File.join(PUBLISHED_DIR, IMAGES_DIR))
end

def generate_meta_tags(front_matter, base_url)
  tags = []
  tags << "<meta property=\"og:title\" content=\"#{front_matter['title']}\" />" if front_matter['title']
  tags << "<meta property=\"og:author\" content=\"#{front_matter['author']}\" />" if front_matter['author']
  tags << "<meta property=\"og:description\" content=\"#{front_matter['summary']}\" />" if front_matter['summary']
  if front_matter['image']
    normalized_image = front_matter['image'].sub(%r{^/}, "")
    tags << "<meta property=\"og:image\" content=\"#{absolute_path(base_url, normalized_image)}\" />"
  end
  tags << "<meta property=\"og:date\" content=\"#{front_matter['date']}\" />" if front_matter['date']
  tags.join("\n")
end


def generate_menu(front_matter, base_url)
  menu_items = front_matter['menu'] || []
  return '' if menu_items.empty?

  items_html = menu_items.map do |entry|
    formatted_path = entry['path'].sub(/\.md$/, '.html')
    if entry['children']
      # Bootstrap dropdown for nested menus
      <<~HTML
        <li class="nav-item dropdown">
          <a class="nav-link dropdown-toggle" href="##{absolute_path(base_url, formatted_path)}" id="navbarDropdown" role="button" data-bs-toggle="dropdown" aria-expanded="false">
            #{entry['name']}
          </a>
          <ul class="dropdown-menu" aria-labelledby="navbarDropdown">
            #{entry['children'].map { |child| "<li><a class='dropdown-item' href='#{absolute_path(base_url, child['path'])}'>#{child['name']}</a></li>" }.join}
          </ul>
        </li>
      HTML
    else
      # Standard navigation item
      <<~HTML
        <li class="nav-item">
          <a class="nav-link" href="#{absolute_path(base_url, formatted_path)}">#{entry['name']}</a>
        </li>
      HTML
    end
  end.join

  # Wrap items in a Bootstrap navigation container
  <<~HTML
    <ul class="navbar-nav me-auto mb-2 mb-lg-0">
      #{items_html}
    </ul>
  HTML
end


def generate_post_list(folder_name, base_url)
  folder_path = File.join(MARKDOWN_DIR, folder_name)
  return "" unless Dir.exist?(folder_path)

  # Group posts by year
  posts_by_year = Dir.glob(File.join(folder_path, "*.md")).each_with_object(Hash.new { |h, k| h[k] = [] }) do |file, years|
    content = File.read(file)
    front_matter, _ = parse_front_matter(content, {})
    next unless front_matter["date"]

    year = Date.parse(front_matter["date"]).year
    file_name = File.basename(file, ".md")
    display_name = file_name.gsub(/[-_]/, " ").split.map(&:capitalize).join(" ")
    link_path = absolute_path(base_url, "#{folder_name}/#{file_name}.html")
    years[year] << { name: display_name, link: link_path }
  end

  # Generate HTML
  posts_by_year.keys.sort.reverse.map do |year|
    post_items = posts_by_year[year].sort_by { |post| post[:name] }.map do |post|
      <<~HTML
        <li class="list-group-item">
          <a class="text-decoration-none" href="#{post[:link]}">#{post[:name]}</a>
        </li>
      HTML
    end.join("\n")

    <<~HTML
      <div class="mb-4">
        <h4 class="fw-bold text-primary">#{year}</h4>
        <ul class="list-group">
          #{post_items}
        </ul>
      </div>
    HTML
  end.join("\n")
end

def generate_rss_feed(base_url)
  rss_file = File.join(PUBLISHED_DIR, "rss.xml")
  defaults = load_defaults
  all_posts = []

  # Collect posts from all directories
  Dir.glob(File.join(MARKDOWN_DIR, "**/*.md")).each do |file|
    content = File.read(file)
    front_matter, _ = parse_front_matter(content, defaults)
    next unless front_matter['date'] && front_matter['title']

    post_date = Date.parse(front_matter['date'])
    post_title = front_matter['title']
    post_summary = front_matter['summary'] || ''
    relative_path = Pathname.new(file).relative_path_from(Pathname.new(MARKDOWN_DIR)).to_s
    post_link = absolute_path(base_url, relative_path.sub(/\.md$/, '.html'))

    all_posts << {
      title: post_title,
      date: post_date,
      summary: post_summary,
      link: post_link
    }
  end

  # Sort posts by date (newest first)
  all_posts.sort_by! { |post| -post[:date].to_time.to_i }

  # Generate RSS feed
  rss = RSS::Maker.make("atom") do |maker|
    maker.channel.author = defaults['author'] || "Krems"
    maker.channel.updated = all_posts.first[:date].to_time if all_posts.any?
    maker.channel.about = absolute_path(base_url, "rss.xml")
    maker.channel.title = defaults['title'] || "RSS Feed"
    maker.channel.link = base_url
    maker.channel.description = defaults['summary'] || "Latest posts"

    all_posts.each do |post|
      maker.items.new_item do |item|
        item.link = post[:link]
        item.title = post[:title]
        item.summary = post[:summary]
        item.updated = post[:date].to_time
      end
    end
  end

  # Write RSS feed to file
  File.write(rss_file, rss.to_s)
  puts "Generated RSS feed at #{rss_file}"
end

def replace_custom_handlebars(content, base_url)
  content.gsub(/\{\{\s*list_posts\(([^)]+)\)\s*\}\}/) { generate_post_list($1.strip, base_url) }
end

def generate_footer(base_url)
  <<~HTML
    <footer class="bg-light py-4 mt-auto">
      <div class="container text-center">
        <p class="mb-2">&copy; #{Time.now.year} | Created with <a href="https://github.com/mreider/krems">Krems</a></p>
        <p class="mb-0">
          <a href="#top" class="text-decoration-none text-primary">Back to top</a>
        </p>
      </div>
    </footer>
  HTML
end

def wrap_body_content(rendered_body_content)
  <<~HTML
    <div class="container py-4">
      #{rendered_body_content}
    </div>
  HTML
end


def convert_markdown_to_html(base_url)
  defaults = load_defaults
  renderer = Redcarpet::Render::HTML.new(hard_wrap: true)
  markdown = Redcarpet::Markdown.new(renderer, tables: true, autolink: true, fenced_code_blocks: true)

  Dir.glob(File.join(MARKDOWN_DIR, "**/*.md")).each do |file|
    relative_path = Pathname.new(file).relative_path_from(Pathname.new(MARKDOWN_DIR)).to_s
    output_file = File.join(PUBLISHED_DIR, relative_path.sub(/\.md$/, ".html"))
    FileUtils.mkdir_p(File.dirname(output_file))

    md_content = File.read(file)
    front_matter, raw_body_content = parse_front_matter(md_content, defaults)

    # Render markdown content and process links
    rendered_body_content = markdown.render(raw_body_content)
    rendered_body_content = convert_links_to_html(rendered_body_content, base_url)
    rendered_body_content = replace_custom_handlebars(rendered_body_content, base_url)

    # Post-specific meta details
    post_title = front_matter['title'] || ''
    post_author = front_matter['author'] || ''
    post_date = front_matter['date'] ? Date.parse(front_matter['date']).strftime('%B %d, %Y') : nil
    post_summary = front_matter['summary'] || ''
    post_image = front_matter['image'] || ''

    # Bootstrap-styled post meta information
    post_meta = <<~HTML
      <div class="mb-4" id="top">
        #{post_image.to_s.strip.empty? ? '' : "<div class='mb-3'><img src='#{absolute_path(base_url, post_image)}' alt='#{post_title}' class='img-fluid rounded'></div>"}
        #{post_title.to_s.strip.empty? ? '' : "<h1 class='h3 text-primary mb-2'>#{post_title}</h1>"}
        #{post_summary.to_s.strip.empty? ? '' : "<p class='text-muted'>#{post_summary}</p>"}
        #{post_author.to_s.strip.empty? && post_date.nil? ? '' : "<p class='text-secondary small'>#{post_author.to_s.strip.empty? ? '' : "By #{post_author}"}#{post_date ? " on #{post_date}" : ''}</p>"}
      </div>
    HTML

    # Generate header, meta tags, and footer
    header = generate_header(base_url, front_matter)
    meta_tags = generate_meta_tags(front_matter, base_url)
    static_assets = generate_static_asset_links(base_url)
    footer = generate_footer(base_url)

    # Combine everything into Bootstrap layout
    html_content = <<~HTML
      <!DOCTYPE html>
      <html lang="en">
      <head>
        <title>#{post_title}</title>
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        #{static_assets}
        #{meta_tags}
      </head>
      <body>
        <div class="wrapper d-flex flex-column min-vh-100">
          <header class="mb-4">
            #{header}
          </header>
          <main class="flex-grow-1">
            <div class="container py-4">
              #{post_meta}
              <div class="content">
                #{rendered_body_content}
              </div>
            </div>
          </main>
          <footer>
            #{footer}
          </footer>
        </div>
        <script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.bundle.min.js"></script>
      </body>
      </html>
    HTML

    # Write the generated HTML to the output file
    File.write(output_file, html_content)
  end
end

def generate_site(local)
  base_url = load_base_url(local)
  clean_published_directory
  unless File.exist?(File.join(MARKDOWN_DIR, "index.md"))
    raise "Error: 'index.md' is missing in the #{MARKDOWN_DIR} directory."
  end
  convert_markdown_to_html(base_url)
  copy_static_assets
  generate_rss_feed(base_url)
end

options = { mode: 'build' }
OptionParser.new do |opts|
  opts.banner = "Usage: ruby krems.rb [options]"
  opts.on("--init", "Initialize a new Krems project") { options[:mode] = 'init' }
  opts.on("--serve", "Run in serve mode (local preview)") { options[:mode] = 'serve' }
  opts.on("--build", "Run in build mode (default)") { options[:mode] = 'build' }
end.parse!

case options[:mode]
when 'init'
  initialize_project
when 'serve'
  require 'sinatra'
  base_url = load_base_url(true)
  generate_site(true)
  set :public_folder, PUBLISHED_DIR
  listen_paths = [MARKDOWN_DIR, CSS_DIR, IMAGES_DIR]
  listener = Listen.to(*listen_paths, only: /\.(md|css|png|jpg|jpeg|gif|svg)$/) { generate_site(true) }
  listener.start
  get('/') { send_file File.join(PUBLISHED_DIR, "index.html") }
  Sinatra::Application.run!
else
  generate_site(false)
end

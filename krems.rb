require 'fileutils'
require 'redcarpet'
require 'pathname'
require 'toml-rb'
require 'date'
require 'titleize'
require 'listen'
require 'optparse'

# Directories
MARKDOWN_DIR = "markdown"
CSS_DIR = "css"
IMAGES_DIR = "images"
PUBLISHED_DIR = "published"
CONFIG_FILE = "config.toml"

# Default content for generated index.md
DEFAULT_INDEX_CONTENT = <<~MARKDOWN
  +++
  title = "Welcome to Krems"
  author = "Krems"
  date = "#{Time.now.strftime('%Y-%m-%d')}"
  summary = "Krems is running, but there was no index.md page, so I generated this one instead 😊."
  +++

  # Welcome to Krems

  Krems is running, but there was no `index.md` page, so I generated this one instead. 😊
MARKDOWN

# Utility: Normalize base URL
def normalize_url(url)
  url = url.to_s.strip
  url.end_with?("/") ? url : "#{url}/"
end

# Load base URL from config.toml
def load_base_url(local = false)
  if local
    "http://127.0.0.1:4567/"
  elsif File.exist?(CONFIG_FILE)
    normalize_url(TomlRB.load_file(CONFIG_FILE)['url'] || "/")
  else
    "/"
  end
end

# Load CSS file from config.toml
def load_css_file
  if File.exist?(CONFIG_FILE)
    config = TomlRB.load_file(CONFIG_FILE)
    config['css'] || "styles.css"
  else
    "styles.css"
  end
end

def clean_published_directory
  puts "Cleaning published directory..."
  FileUtils.rm_rf(PUBLISHED_DIR)
  FileUtils.mkdir_p(PUBLISHED_DIR)
  puts "Published directory cleaned."
end

def ensure_index_md
  index_file = File.join(MARKDOWN_DIR, "index.md")
  unless File.exist?(index_file)
    puts "No index.md found. Creating default index.md..."
    File.write(index_file, DEFAULT_INDEX_CONTENT)
    puts "Default index.md created."
  end
end

def load_defaults
  defaults_file = "defaults.toml"
  File.exist?(defaults_file) ? TomlRB.load_file(defaults_file) : {}
end

def merge_defaults(front_matter, defaults)
  defaults.each do |key, value|
    front_matter[key] = value unless front_matter.key?(key)
  end
  front_matter
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
  relative_path = relative_path.sub(%r{^/}, "") # Remove leading slash if present
  "#{base_url}#{relative_path}"
end

def convert_links_to_html(content, base_url)
  content.gsub(/href="(\/?[a-zA-Z0-9\-_\/\.]+)\.md"/) do
    link = $1
    "href=\"#{absolute_path(base_url, "#{link}.html")}\""
  end
end

def update_image_links(content, base_url)
  content.gsub(/!\[([^\]]*)\]\((\/?images\/[^\)]+)\)/) do
    alt_text, image_path = $1, $2
    "![#{alt_text}](#{absolute_path(base_url, image_path)})"
  end
end






def copy_static_assets
  puts "Copying static assets..."
  FileUtils.mkdir_p(File.join(PUBLISHED_DIR, CSS_DIR))
  FileUtils.mkdir_p(File.join(PUBLISHED_DIR, IMAGES_DIR))

  FileUtils.cp_r(Dir.glob(File.join(CSS_DIR, "*")), File.join(PUBLISHED_DIR, CSS_DIR))
  FileUtils.cp_r(Dir.glob(File.join(IMAGES_DIR, "*")), File.join(PUBLISHED_DIR, IMAGES_DIR))

  puts "Static assets copied successfully."
end

def generate_meta_tags(front_matter, base_url)
  tags = []
  tags << "<meta property=\"og:title\" content=\"#{front_matter['title']}\" />" if front_matter['title']
  tags << "<meta property=\"og:author\" content=\"#{front_matter['author']}\" />" if front_matter['author']
  tags << "<meta property=\"og:description\" content=\"#{front_matter['summary']}\" />" if front_matter['summary']
  if front_matter['image']
    # Ensure the image path is relative to the base URL
    normalized_base = normalize_url(base_url)
    normalized_image = front_matter['image'].sub(%r{^/}, "") # Remove leading slash if present
    tags << "<meta property=\"og:image\" content=\"#{normalized_base}#{normalized_image}\" />"
  end
  tags << "<meta property=\"og:date\" content=\"#{front_matter['date']}\" />" if front_matter['date']
  tags.join("\n")
end


def generate_static_asset_links(base_url)
  <<~HTML
    <link rel="stylesheet" href="#{absolute_path(base_url, "css/#{load_css_file}")}">
    <link rel="icon" href="#{absolute_path(base_url, "images/favicon.ico")}">
  HTML
end

def generate_menu(front_matter, base_url)
  return "" unless front_matter["menu"]

  menu_items = front_matter["menu"].map do |entry|
    formatted_path = entry["path"].gsub(/^\//, "").sub(/\.md$/, ".html")
    display_name = entry["name"].gsub(/([a-z])([A-Z])/, '\1 \2').titleize # Convert PascalCase to spaced words
    "<td><h4><a href=\"#{base_url}#{formatted_path}\">#{display_name}</a></h4></td>"
  end
  "<table><tr>#{menu_items.join('<td><h4>•</h4></td>')}</tr></table>"
end


def generate_post_list(folder_name, base_url)
  folder_path = File.join(MARKDOWN_DIR, folder_name)
  return "" unless Dir.exist?(folder_path)

  posts_by_year = Dir.glob(File.join(folder_path, "*.md")).each_with_object(Hash.new { |h, k| h[k] = [] }) do |file, years|
    content = File.read(file)
    front_matter, _ = parse_front_matter(content, {})
    next unless front_matter["date"]

    year = Date.parse(front_matter["date"]).year
    file_name = File.basename(file, ".md")
    display_name = file_name.gsub(/[-_]/, " ").split.map(&:capitalize).join(" ") # Convert to title case
    link_path = absolute_path(base_url, "#{folder_name}/#{file_name}.html")
    years[year] << { name: display_name, link: link_path }
  end

  posts_by_year.keys.sort.reverse.map do |year|
    <<~HTML
      <h4>#{year}</h4>
      <ul>
        #{posts_by_year[year].sort_by { |post| post[:name] }.map { |post| "<li><a href=\"#{post[:link]}\">#{post[:name]}</a></li>" }.join("\n")}
      </ul>
    HTML
  end.join("\n")
end


def replace_custom_handlebars(content, base_url)
  content.gsub(/\{\{\s*list_posts\(([^)]+)\)\s*\}\}/) do
    folder_name = $1.strip
    generate_post_list(folder_name, base_url)
  end
end

def generate_footer(base_url)
  <<~HTML
    <footer>
      <h4>&copy; #{Time.now.year} Matt Reider &bull; Site generated with <a href="#{base_url}">Krems</a></h4>
    </footer>
  HTML
end

def convert_markdown_to_html(base_url)
  puts "Base URL used for generation: #{base_url}" # Debugging output
  defaults = load_defaults
  renderer = Redcarpet::Render::HTML.new
  markdown = Redcarpet::Markdown.new(renderer, tables: true, autolink: true, fenced_code_blocks: true)

  puts "Converting Markdown files to HTML..."
  Dir.glob(File.join(MARKDOWN_DIR, "**/*.md")).each do |file|
    relative_path = Pathname.new(file).relative_path_from(Pathname.new(MARKDOWN_DIR)).to_s
    output_file = File.join(PUBLISHED_DIR, relative_path.sub(/\.md$/, ".html"))
    FileUtils.mkdir_p(File.dirname(output_file))

    md_content = File.read(file)
    front_matter, body_content = parse_front_matter(md_content, defaults)
    body_content = markdown.render(body_content)

    # Apply link and image updates
    body_content = update_image_links(body_content, base_url)
    body_content = convert_links_to_html(body_content, base_url)
    body_content = replace_custom_handlebars(body_content, base_url)

    menu = generate_menu(front_matter, base_url)
    meta_tags = generate_meta_tags(front_matter, base_url)
    static_assets = generate_static_asset_links(base_url)
    footer = generate_footer(base_url)

    header_content = ""
    header_content << "<h1>#{front_matter['title']}</h1>" if front_matter['title']
    if front_matter['author'] && front_matter['date']
      formatted_date = Date.parse(front_matter['date']).strftime("%B %d, %Y")
      header_content << "<h2>By #{front_matter['author']} on #{formatted_date}</h2>"
    end
    header_content << "<h3>#{front_matter['summary']}</h3>" if front_matter['summary']

    # Normalize the base URL for the image in the front matter
    if front_matter["image"]
      normalized_base = normalize_url(base_url)
      normalized_image = front_matter["image"].sub(%r{^/}, "") # Remove leading slash if present
      header_content << <<~HTML
        <img src="#{normalized_base}#{normalized_image}" alt="#{front_matter['title']}">
      HTML
    end

    File.write(output_file, <<~HTML)
      <!DOCTYPE html>
      <html>
      <head>
        <title>#{front_matter['title'] || 'Krems'}</title>
        <meta name="viewport" content="width=device-width, initial-scale=1.0">
        #{static_assets}
        #{meta_tags}
      </head>
      <body>
        #{menu}
        #{header_content}
        #{body_content}
        #{footer}
      </body>
      </html>
    HTML
    puts "Generated: #{output_file}"
  end
  puts "Markdown to HTML conversion complete."
end


def generate_site(local)
  base_url = load_base_url(local)
  clean_published_directory
  ensure_index_md
  convert_markdown_to_html(base_url)
  copy_static_assets
  puts "Site generation complete."
end

options = { mode: 'build' }
OptionParser.new do |opts|
  opts.banner = "Usage: ruby krems.rb [options]"

  opts.on("--serve", "Run in serve mode (local preview)") do
    options[:mode] = 'serve'
  end

  opts.on("--build", "Run in build mode (default)") do
    options[:mode] = 'build'
  end
end.parse!

if options[:mode] == 'serve'
  require 'sinatra'
  base_url = load_base_url(true)
  puts "Starting local server..."
  generate_site(true)

  set :public_folder, PUBLISHED_DIR

  # Watch for file changes
  listen_paths = [MARKDOWN_DIR, CSS_DIR, IMAGES_DIR]
  listener = Listen.to(*listen_paths, only: /\.(md|css|png|jpg|jpeg|gif|svg)$/) do
    puts "Change detected! Rebuilding site..."
    generate_site(true)
    puts "Rebuild complete. Refresh your browser."
  end
  listener.start

  get '/' do
    send_file File.join(PUBLISHED_DIR, "index.html")
  end

  Sinatra::Application.run!
else
  puts "Building site for deployment..."
  generate_site(false)
end
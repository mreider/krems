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
  url = url.to_s
  url.end_with?('/') ? url : "#{url}/"
end

# Load base URL from config.toml
def load_base_url(local = false, port = 4567)
  if local
    "http://127.0.0.1:#{port}/"
  elsif ENV['GITHUB_ACTIONS'] || File.exist?(CONFIG_FILE)
    url = TomlRB.load_file(CONFIG_FILE)['url'] rescue "/"
    url.end_with?("/") ? url : "#{url}/"
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

def convert_links_to_html(content, base_url)
  normalized_base = normalize_url(base_url)

  content.gsub(/href="(\/?[a-zA-Z0-9\-_\/\.]+)\.md"/) do
    link = $1
    # Ensure no double slashes by stripping leading slash
    absolute_path = link.start_with?("/") ? link[1..] : link
    "href=\"#{normalized_base}#{absolute_path}.html\""
  end
end

def update_image_links(content, base_url)
  # Ensure base URL ends with a single slash
  normalized_base = normalize_url(base_url)

  content.gsub(/!\[([^\]]*)\]\((\/?images\/[^\)]+)\)/) do
    alt_text, image_path = $1, $2
    # Ensure no double slashes by stripping the leading slash from image_path
    normalized_path = image_path.sub(%r{^/}, "")
    "![#{alt_text}](#{normalized_base}#{normalized_path})"
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
    normalized_image = front_matter['image'].sub(%r{^/}, "")
    tags << "<meta property=\"og:image\" content=\"#{base_url}#{normalized_image}\" />"
  end
  tags << "<meta property=\"og:date\" content=\"#{front_matter['date']}\" />" if front_matter['date']
  tags.join("\n")
end

def generate_static_asset_links(base_url)
  normalized_base = normalize_url(base_url)
  css_file = load_css_file

  <<~HTML
    <link rel="stylesheet" href="#{normalized_base}css/#{css_file}">
    <link rel="apple-touch-icon" sizes="180x180" href="#{normalized_base}images/apple-touch-icon.png">
    <link rel="icon" type="image/png" sizes="32x32" href="#{normalized_base}images/favicon-32x32.png">
    <link rel="icon" type="image/png" sizes="16x16" href="#{normalized_base}images/favicon-16x16.png">
    <link rel="manifest" href="#{normalized_base}images/site.webmanifest">
    <link rel="icon" href="#{normalized_base}images/favicon.ico">
    <meta name="theme-color" content="#ffffff">
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
    display_name = File.basename(file, ".md").titleize
    link_path = "#{base_url}#{folder_name}/#{File.basename(file, '.md')}.html"
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
    body_content = update_image_links(body_content, base_url) # Fix image links
    body_content = convert_links_to_html(body_content, base_url) # Fix internal links
    body_content = replace_custom_handlebars(body_content, base_url)

    menu = generate_menu(front_matter, base_url)
    meta_tags = generate_meta_tags(front_matter, base_url)
    static_assets = generate_static_asset_links(base_url)
    footer = generate_footer(base_url)

    header_content = front_matter['title'] ? "<h1>#{front_matter['title']}</h1>" : ""

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

def generate_site(base_url)
  clean_published_directory
  ensure_index_md
  convert_markdown_to_html(base_url)
  copy_static_assets
  puts "Site generation complete."
end

options = { mode: 'build' }
OptionParser.new do |opts|
  opts.banner = "Usage: ruby krems.rb [options]"

  opts.on("--serve", "Run in serve mode") do
    options[:mode] = 'serve'
  end

  opts.on("--build", "Run in build mode (default)") do
    options[:mode] = 'build'
  end
end.parse!

if options[:mode] == 'serve'
  require 'sinatra'
  base_url = load_base_url(true)
  puts "Starting site generation for local testing..."
  generate_site(base_url)

  set :public_folder, PUBLISHED_DIR

  # Watch for changes and rebuild the site
  listen_paths = [MARKDOWN_DIR, CSS_DIR, IMAGES_DIR] # Exclude config.toml
  listener = Listen.to(*listen_paths, only: /\.(md|css|png|jpg|jpeg|gif|svg)$/) do |modified, added, removed|
    puts "Change detected! Files modified: #{modified.join(', ')}, added: #{added.join(', ')}, removed: #{removed.join(', ')}"
    puts "Rebuilding site..."
    generate_site(load_base_url(true))
    puts "Rebuild complete. Refresh the browser to see the changes."
  end

  listener.start

  get '/' do
    send_file File.join(PUBLISHED_DIR, "index.html")
  end

  Sinatra::Application.run!
else
  base_url = load_base_url
  puts "Starting site generation..."
  generate_site(base_url)
end
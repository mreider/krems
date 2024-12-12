require 'fileutils'
require 'redcarpet'
require 'pathname'
require 'toml-rb'
require 'date'
require 'titleize'
require 'listen'
require 'optparse'

MARKDOWN_DIR = "markdown"
CSS_DIR = "css"
IMAGES_DIR = "images"
PUBLISHED_DIR = "published"
CONFIG_FILE = "config.toml"

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

def normalize_url(url)
  url = url.to_s.strip
  url.end_with?("/") ? url : "#{url}/"
end

def initialize_project
  puts "Initializing new Krems project..."

  [MARKDOWN_DIR, CSS_DIR, IMAGES_DIR, PUBLISHED_DIR].each do |dir|
    FileUtils.mkdir_p(dir)
    puts "Created directory: #{dir}"
  end

  config_content = <<~TOML
    url = "http://127.0.0.1:4567/"
    css = "styles.css"
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

  default_css_content = <<~CSS
    body {
      font-family: Helvetica, Arial, sans-serif;
      line-height: 1.5;
      margin: 0;
      padding: 20px;
    }

    h1, h2, h3, h4, h5, h6 {
      color: #333;
    }

    a {
      color: #3973ad;
      text-decoration: none;
    }

    a:hover {
      color: #4183C4;
    }
  CSS
  File.write(File.join(CSS_DIR, "styles.css"), default_css_content)
  puts "Created file: css/styles.css"

  puts "Krems project initialized successfully."
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

def load_css_file
  if File.exist?(CONFIG_FILE)
    config = TomlRB.load_file(CONFIG_FILE)
    config['css'] || 'styles.css'
  else
    'styles.css'
  end
end


def clean_published_directory
  FileUtils.rm_rf(PUBLISHED_DIR)
  FileUtils.mkdir_p(PUBLISHED_DIR)
end

def ensure_index_md
  index_file = File.join(MARKDOWN_DIR, "index.md")
  unless File.exist?(index_file)
    File.write(index_file, DEFAULT_INDEX_CONTENT)
  end
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
    parsed_front_matter = TomlRB.parse(front_matter)
    merged_front_matter = merge_defaults(parsed_front_matter, defaults)
    [merged_front_matter, body.strip]
  else
    [merge_defaults({}, defaults), content.strip] # Use defaults if no front matter
  end
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
  content.gsub(/href="(\/?[a-zA-Z0-9\-_\/\.]+)\.md"/) { "href=\"#{absolute_path(base_url, "#{$1}.html")}\"" }
end

def update_image_links(content, base_url)
  content.gsub(/!\[([^\]]*)\]\((\/?images\/[^\)]+)\)/) do
    alt_text, image_path = $1, $2
    "![#{alt_text}](#{absolute_path(base_url, image_path)})"
  end
end

def generate_static_asset_links(base_url)
  css_file = load_css_file
  "<link rel='stylesheet' href='#{absolute_path(base_url, "css/#{css_file}")}'>" \
  "<link rel='icon' href='#{absolute_path(base_url, 'images/favicon.ico')}'>"
end

def generate_header(base_url, front_matter)
  title = front_matter['title'] || ''
  summary = front_matter['summary'] || ''
  date = front_matter['date']
  author = front_matter['author'] || ''

  # Generate a pretty date if available
  pretty_date = Date.parse(date).strftime('%B %d, %Y') rescue nil if date

  # Generate the author and date HTML
  author_date_html = if author.strip.empty? && pretty_date.nil?
                       ''
                     elsif !author.strip.empty? && pretty_date
                       "<p class='author-date'>By #{author} on #{pretty_date}</p>"
                     elsif !author.strip.empty?
                       "<p class='author-date'>By #{author}</p>"
                     elsif pretty_date
                       "<p class='author-date'>On #{pretty_date}</p>"
                     end

  # Generate the menu HTML
  menu_html = generate_menu(front_matter, base_url)

  # Combine into the header HTML
  "<header id='header' class='py4'>" +
    "<a href='#{base_url}' class='flex items-center'>" +
    "<div id='logo' style='background-image: url(\"#{absolute_path(base_url, 'images/logo.png')}\"); width: 50px; height: 50px; background-size: cover; background-position: center;'></div>" +
    "<div id='title' class='ml2'>" +
    "<h1>#{title}</h1>" +
    "</div></a>" +
    "#{menu_html}" +
    "#{author_date_html}" +
    (summary.strip.empty? ? '' : "<p class='summary'>#{summary}</p>") +
  "</header>"
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
  menu_items = front_matter['menu'] || [] # Ensure menu is always an array
  return '' if menu_items.empty? # Skip menu generation if no items

  items_html = menu_items.map do |entry|
    formatted_path = entry['path'].sub(/\.md$/, '.html')
    "<li><a href='#{absolute_path(base_url, formatted_path)}'>#{entry['name']}</a></li>"
  end.join("\n")

  "<div id='nav'><ul>#{items_html}</ul></div>"
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
    display_name = file_name.gsub(/[-_]/, " ").split.map(&:capitalize).join(" ")
    link_path = absolute_path(base_url, "#{folder_name}/#{file_name}.html")
    years[year] << { name: display_name, link: link_path }
  end

  posts_by_year.keys.sort.reverse.map do |year|
    post_items = posts_by_year[year].sort_by { |post| post[:name] }.map do |post|
      "<li class='post-item my1'><a class='post-link' href='#{post[:link]}'>#{post[:name]}</a></li>"
    end.join

    "<h4 class='h3 my2'>#{year}</h4><ul class='post-list list-none'>#{post_items}</ul>"
  end.join
end

def replace_custom_handlebars(content, base_url)
  content.gsub(/\{\{\s*list_posts\(([^)]+)\)\s*\}\}/) { generate_post_list($1.strip, base_url) }
end

def generate_footer(base_url)
  "<footer class=\"footer px3 py2 border-box\">" \
  "&copy; #{Time.now.year} Matt Reider | Created with <a href='https://github.com/mreider/krems'>Krems</a>" \
  "</footer>"
end

def convert_markdown_to_html(base_url)
  defaults = load_defaults
  renderer = Redcarpet::Render::HTML.new(hard_wrap: false)
  markdown = Redcarpet::Markdown.new(renderer, tables: true, autolink: true, fenced_code_blocks: true)
  

  Dir.glob(File.join(MARKDOWN_DIR, "**/*.md")).each do |file|
    relative_path = Pathname.new(file).relative_path_from(Pathname.new(MARKDOWN_DIR)).to_s
    output_file = File.join(PUBLISHED_DIR, relative_path.sub(/\.md$/, ".html"))
    FileUtils.mkdir_p(File.dirname(output_file))

    md_content = File.read(file)
    front_matter, body_content = parse_front_matter(md_content, defaults)
    body_content = markdown.render(body_content)
    body_content = update_image_links(body_content, base_url)
    body_content = convert_links_to_html(body_content, base_url)
    body_content = replace_custom_handlebars(body_content, base_url)

    header = generate_header(base_url, front_matter)
    meta_tags = generate_meta_tags(front_matter, base_url)
    static_assets = generate_static_asset_links(base_url)
    footer = generate_footer(base_url)

    html_content = "<!DOCTYPE html>" \
    "<html>" \
    "<head>" \
    "<title>#{front_matter['title'] || 'krems'}</title>" \
    "<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">" \
    "#{static_assets}" \
    "#{meta_tags}" \
    "</head>" \
    "<body class=\"content max-width mx-auto\">" \
    "#{header}" \
    "<main class=\"mx3 my3\">" \
    "#{body_content}" \
    "</main>" \
    "#{footer}" \
    "</body>" \
    "</html>"


    File.write(output_file, html_content)
  end
end




def generate_site(local)
  base_url = load_base_url(local)
  clean_published_directory
  ensure_index_md
  convert_markdown_to_html(base_url)
  copy_static_assets
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

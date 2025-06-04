# Krems

Krems lets you publish Markdown websites to Github pages.

[☕️ Buy me a coffee ☕️](https://coff.ee/mreider)

## Getting started

1. Fork the [example site](https://github.com/mreider/krems-example/)
2. Enable the workflow
3. Turn on Github Pages in the repository
    - Push from the gh-pages branch
    - Use the /(root) folder
4. View your repository's actions
    - An action should be running
    - When it's done your website will be ready

You can view your website in a browser.

For example:

https://your-gh-user.github.io/krems-example/

Note: The config.yaml file contains my example URL, so links will redirect to that URL instead of yours. To fix this, edit your local config.yaml file and redeploy.

## Learn from the example and build your own site

The example site shows all of the functionality of Krems. The stylesheet is fixed and generic for everyone. All Krems sites look the same. If you want to improve it, open a pull request and I can update it.

If you want to change the CSS in a live view it's [here](https://codepen.io/matthew-reider/pen/dPoOebJ).

Note: I'll be adding custom CSS soon.

## Images

You must store your images in an /images folder and reference them using normal markdown. You can have subfolders of images to keep them organized.

## Page Types

There are two page types.

## List pages

- show links to other pages
- only show pages that have dates
- (usually) have no markdown content
- (usually) show a list of pages in a single directory
- (usually) exist as index.md in a directory
- have the following front matter:

```
---
title: "Krems Home Page"
type: list
created: 2025-06-04T09:24
updated: 2025-06-04T09:39
---
```

## List page filters

List page filters expand the functionality of a list page

- shows all pages in all subdirectories with:
    - specific tags (or...)
    - specific authors
- have the following front matter:


```
---
title: Krems Home Page
type: list
tagFilter:
  - about
authorFilter:
  - Matt
---
```

## Default pages

- have Markdown content
- include an (optional) image
    - is converted to an Open Graph image
    - displayed as a preview images when someone shares the page URL
- have the following frontmatter:

```
---
title: "Krems City Info"
date: "2024-11-26"
image: "/images/krems1.png"
author: "Matt"
tags: ["about"]
---
```

## About config.yaml

- required at root directory
- must have `basePath` if home page is in a subdirectory
- must have `devPath` to run locally without subdirectory
- follows example below:

```
website:
  url: "https://mreider.github.io/krems-example"
  name: "Krems Example Site"
  basePath: "/krems-example"
  devPath: "/"

menu:
  - title: "Home"
    path: "index.md"
  - title: "Universities"
    path: "universities/index.md"
```


## Running Krems locally

1. Download the [latest binary](https://github.com/mreider/krems/releases) and put it in your path
2. Create a Krems site or clone the [example](https://github.com/mreider/krems-example)
3. Run and browse the site:
    - `krems --run`
    - runs at localhost:8080 (--port to override)
4. this creates a .tmp directory with HTML
5. clean the .tmp directory using:
    - `krems --clean`
6. to build the site without running:
    - `krems --build`

## About the Github Action

The [example](https://github.com/mreider/krems-example) has a Workflow that uses the [Krems Github Action](https://github.com/mreider/krems-deploy-action).

- Generates the site on push
- Commits the website to gh-pages branch using the /(root) directory
- Creates a CNAME file if the config.yaml has a custom domain

## Questions / feedback

- [about Krems static site generation](https://github.com/mreider/krems/issues)
- [about the Krems Obsidian plugin](https://github.com/mreider/krems-obsidian-plugin/issues)
- [about the Krems Github Action](https://github.com/mreider/krems-deploy-action/issues)

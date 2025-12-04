---
publish: true
---

# General Configuration

The configuration settings for Blaze can be found in `blaze.config.json`. Below are the details for each option:

- `pageTitle` The main title of your website. This is also used as the site name when generating the RSS Feed.

- `pageTitleSuffix` A text string appended to the end of the page title. **Note:** This only affects the title visible in the **browser tab**, not the main title displayed at the top of the page content.

- `locale` Defines the region and language settings. This is used for internationalization (i18n) and to determine the correct date formatting.

- `baseURL` The absolute URL where your site is hosted. This is critical for Sitemaps and RSS feeds to function correctly. **Important:** Enter the domain only (e.g., `blaze.rhmt.my.id`). **Do not** include the protocol (e.g., `https://`) or any leading/trailing slashes.

- `ignorePatterns` A list of glob patterns used to exclude specific files or folders. Blaze will skip these files when scanning the content folder (useful for keeping pages private).

- `publishMode` Controls the publication logic. If set to `explicit`, a document will **not** be published unless you manually add the `publish: true` property to the document's frontmatter.

**Note:** Configuration changes are automatically detected during development server (`serve` mode) and will trigger a rebuild without needing to restart the server or recompile the binary.

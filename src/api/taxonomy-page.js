import imgproxyURL from "@/api/imgproxy-url.js";
import { toRel } from "@/utils/link.js";
import maybeDate from "@/utils/maybe-date.js";

export class TaxonomyPage {
  constructor(data) {
    this.id = data["id"] ?? "";
    this.body = data["body"] ?? "";
    this.frontmatter = data["frontmatter"] ?? {};
    this.filePath = data["file_path"] ?? "";
    this.urlPath = data["url_path"] ?? "";
    this.lastPublished = maybeDate(data, "last_published");
    this.publicationDate = maybeDate(this.frontmatter, "published");
    this.kicker = this.frontmatter["kicker"] ?? "";
    this.title = this.frontmatter["title"] ?? "";
    this.linkTitle = this.frontmatter["linktitle"] ?? "";
    this.titleTag = this.frontmatter["title-tag"] ?? "";
    this.summary = this.frontmatter["description"] ?? "";
    this.appImage = this.frontmatter["app-image"] ?? "";
    this.appImageGravity = this.frontmatter["app-image-gravity"] ?? "";
    this.appImageDescription = this.frontmatter["app-image-description"] ?? "";
    this.appImageCredit = this.frontmatter["app-image-credit"] ?? "";
    this.image = this.frontmatter["image"] ?? "";
    this.imageGravity = this.frontmatter["image-gravity"] ?? "";
    this.imageDescription = this.frontmatter["image-description"] ?? "";
    this.imageCredit = this.frontmatter["image-credit"] ?? "";
    this.imageSize = this.frontmatter["image-size"] ?? "";
    this.modalExclude = this.frontmatter["modal-exclude"] ?? false;
    this.suppressAds = this.frontmatter["suppress-ads"] ?? false;
    this.noIndex = this.frontmatter["no-index"] ?? null;
    this.overrideURL = this.frontmatter["url"] ?? "";
    this.aliases = this.frontmatter["aliases"] ?? [];
    this.layout = this.frontmatter["layout"] ?? "";
    this.hideDescription = this.frontmatter["hide-description"] ?? false;
    this.descriptionHed = this.frontmatter["description-hed"] ?? "";
    this.descriptionDek = this.frontmatter["description-dek"] ?? "";

    this.shouldUpdateURLPath = false;
  }

  get taxoName() {
    return this.filePath?.replace(
      /^content\/(topics|series)\/(.+)\/_index\.md$/,
      "$2"
    );
  }

  get taxoKind() {
    return this.filePath?.replace(
      /^content\/(topics|series)\/(.+)\/_index\.md$/,
      "$1"
    );
  }

  get taxoLink() {
    if (this.taxoKind === "topics") {
      return { name: "Topic Pages", to: { name: "topic-pages" } };
    }
    return {
      name: "Investigation Series Pages",
      to: { name: "series-pages" },
    };
  }

  get taxoPage() {
    return { name: "taxonomy-page", params: { id: "" + this.id } };
  }

  get isPublished() {
    return !!this.lastPublished;
  }

  get isFutureDated() {
    return this.publicationDate && this.publicationDate > new Date();
  }

  get link() {
    if (this.urlPath) {
      return new URL(this.urlPath, "https://www.spotlightpa.org").href;
    }
    if (this.overrideURL) {
      return new URL(this.overrideURL, "https://www.spotlightpa.org").href;
    }
    return "";
  }

  changeURL() {
    if (!this.isPublished) return;
    let oldURLPath = new URL(this.link).pathname;
    let message = `Are you sure you want to change the URL? Current URL is ${oldURLPath}. Changing the URL will automatically add a redirect from the old URL to a new one. Please enter new URL below.`;
    let newURLPath = window.prompt(message, oldURLPath);
    if (!newURLPath || newURLPath === oldURLPath) return;
    let newURL = new URL(newURLPath, "https://www.spotlightpa.org");
    newURLPath = newURL.pathname;
    this.aliases.push(oldURLPath);
    this.overrideURL = newURLPath;
    this.urlPath = newURLPath;
    this.shouldUpdateURLPath = true;
  }

  getImagePreviewURL(options) {
    if (!this.image || this.image.match(/^http/)) {
      return "";
    }
    return imgproxyURL(this.image, options);
  }

  getAppImagePreviewURL() {
    if (!this.appImage || this.appImage.match(/^http/)) {
      return this.getImagePreviewURL({
        width: 400,
        height: 500,
        gravity: this.imageGravity,
      });
    }
    return imgproxyURL(this.appImage, {
      width: 400,
      height: 500,
      gravity: this.appImageGravity,
    });
  }

  get imagePreviewURL() {
    return this.getImagePreviewURL();
  }

  get thumbnailURL() {
    return imgproxyURL(this.image, {
      width: 256,
      height: 192,
      extension: "webp",
      gravity: this.imageGravity,
    });
  }

  toJSON() {
    return {
      id: this.id,
      set_frontmatter: true,
      frontmatter: {
        // preserve unknown props
        ...this.frontmatter,
        // copy others
        published: this.publicationDate,
        kicker: this.kicker,
        title: this.title,
        linktitle: this.linkTitle,
        "title-tag": this.titleTag,
        description: this.summary,
        "app-image": this.appImage,
        "app-image-gravity": this.appImageGravity,
        "app-image-description": this.appImageDescription,
        "app-image-credit": this.appImageCredit,
        image: this.image,
        "image-gravity": this.imageGravity,
        "image-description": this.imageDescription,
        "image-credit": this.imageCredit,
        "image-size": this.imageSize,
        "modal-exclude": this.modalExclude,
        "suppress-ads": this.suppressAds,
        "no-index": this.noIndex,
        url: toRel(this.overrideURL),
        aliases: this.aliases,
        layout: this.layout,
        "hide-description": this.hideDescription,
        "description-hed": this.descriptionHed,
        "description-dek": this.descriptionDek,
      },
      set_body: true,
      body: this.body,
      set_schedule_for: false,
      url_path: this.shouldUpdateURLPath ? this.urlPath : "",
      set_last_published: false,
    };
  }
}

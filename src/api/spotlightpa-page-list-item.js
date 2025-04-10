import { reactive } from "vue";

export default class PageListItem {
  static from(rawData) {
    let pages = rawData.pages;
    if (!pages) return [];
    return pages.map((page) => reactive(new PageListItem(page)));
  }

  constructor(data) {
    this.id = data["id"] || "";
    this.internalID = data["internal_id"] || "";
    this.title = data["title"] || "";
    this.blurb = data["blurb"] || "";
    this.description = data["description"] || "";
    this.filePath = data["file_path"] || "";
    this.urlPath = data["url_path"] || "";
    this.image = data["image"] || "";
    this.createdAt = PageListItem.getDate(data, "created_at");
    this.publicationDate = PageListItem.getDate(data, "publication_date");
    this.updatedAt = PageListItem.getDate(data, "updated_at");
    this.lastPublished = PageListItem.getDate(data, "last_published");
    this.scheduleFor = PageListItem.getDate(data, "schedule_for");
  }

  static getDate(data, prop) {
    let date = data[prop] ?? null;
    return date && new Date(date);
  }

  get isPublished() {
    return !!this.lastPublished;
  }

  get status() {
    if (this.isPublished) {
      return "pub";
    }
    return this.scheduleFor ? "sked" : "none";
  }

  get link() {
    return {
      name: "news-page",
      params: {
        id: "" + this.id,
      },
    };
  }
}

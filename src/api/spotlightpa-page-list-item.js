import { computed } from "@vue/composition-api";

export default class PageListItem {
  static from(apiState) {
    return computed(() => {
      let pages = apiState.rawData?.pages;
      if (!pages) return [];
      return pages.map((page) => new PageListItem(page));
    });
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
    this.publishedAt = PageListItem.getDate(data, "published_at");
    this.updatedAt = PageListItem.getDate(data, "updated_at");
    this.lastPublished = PageListItem.getNullableDate(data, "last_published");
    this.scheduleFor = PageListItem.getNullableDate(data, "schedule_for");
  }

  static getDate(data, prop) {
    let date = data[prop] ?? null;
    return date && new Date(date);
  }

  static getNullableDate(data, prop) {
    return data[prop]?.Valid ? new Date(data[prop].Time) : null;
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
}

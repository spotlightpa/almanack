import ArcArticle from "./arc-article.js";
import maybeDate from "@/utils/maybe-date.js";

export default class SharedArticle {
  constructor(rawData) {
    this.init(rawData);
  }

  init(data) {
    this["id"] = data["id"] ?? "";
    this["_status"] = data["status"] ?? "";
    this["note"] = data["note"] ?? "";
    this["sourceType"] = data["source_type"] ?? "";
    this["sourceID"] = data["source_id"] ?? "";
    this["rawData"] = data["raw_data"] ?? "";
    this["pageID"] = "" + (data["page_id"]?.Int64 || "");
    this["embargoUntil"] = maybeDate(data, "embargo_until");
    this["createdAt"] = maybeDate(data, "created_at");
    this["updatedAt"] = maybeDate(data, "updated_at");

    this.arc = null;
    if (this.isArc) {
      this.arc = new ArcArticle({
        note: this.note,
        ...data.raw_data,
      });
    }
  }

  get isArc() {
    return this.sourceType === "arc";
  }
  get isUnderEmbargo() {
    if (!this.embargoUntil) {
      return false;
    }
    return new Date() < this.embargoUntil;
  }
  get isPreviewed() {
    return this._status === "P";
  }
  get isShared() {
    return this._status === "S";
  }
  get isReleased() {
    return this.isShared && !this.isUnderEmbargo;
  }
  get status() {
    if (this.isReleased) {
      return "released";
    }
    if (this.isShared) {
      return "embargo";
    }
    if (this.isPreviewed) {
      return "preview";
    }
    return "imported";
  }
  get statusVerbose() {
    return (
      {
        imported: "Imported",
        preview: "Preview Available",
        embargo: "Available Under Embargo",
        released: "Released",
      }[this.status] || "System Error"
    );
  }
  get detailsRoute() {
    return { name: "article", params: { id: this.id } };
  }
  get pageRoute() {
    if (!this.pageID) {
      return null;
    }
    return { name: "news-page", params: { id: this.pageID } };
  }
  get slug() {
    return this.isArc ? this.arc.slug : "TKTK";
  }
}

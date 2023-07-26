import ArcArticle from "./arc-article.js";
import maybeDate from "@/utils/maybe-date.js";

export default class SharedArticle {
  constructor(rawData) {
    this.init(rawData);
  }

  static fromArc(data) {
    return new SharedArticle({
      ...data,
      source_type: "arc",
      source_id: data.arc_id,
      id: data.shared_article_id,
    });
  }

  init(data) {
    this["id"] = data["id"] ?? "";
    this["_status"] = data["status"] ?? "";
    this["note"] = data["note"] ?? "";
    this["sourceType"] = data["source_type"] ?? "";
    this["sourceID"] = data["source_id"] ?? "";
    this["rawData"] = data["raw_data"] ?? "";
    this["pageID"] = "" + (data["page_id"] || "");
    this["internalID"] = data["internal_id"] ?? "";
    this["budget"] = data["budget"] ?? "";
    this["hed"] = data["hed"] ?? "";
    this["description"] = data["description"] ?? "";
    this["blurb"] = data["blurb"] ?? "";
    this["publicationDate"] = maybeDate(data, "publication_date");
    this["embargoUntil"] = maybeDate(data, "embargo_until");
    this["createdAt"] = maybeDate(data, "created_at");
    this["updatedAt"] = maybeDate(data, "updated_at");
    this["lastUpdated"] = maybeDate(data, "last_updated");
    this["byline"] = data["byline"] ?? "";
    this["ledeImage"] = data["lede_image"] ?? "";
    this["ledeImageCredit"] = data["lede_image_credit"] ?? "";
    this["ledeImageDescription"] = data["lede_image_description"] ?? "";
    this["ledeImageCaption"] = data["lede_image_caption"] ?? "";
    if (this.isGDoc) {
      this["gdocs"] = data["gdocs"] ?? {};
      this.gdocs.embeds = this.gdocs.embeds ?? [];
      this.gdocs.warnings = this.gdocs.warnings ?? [];
      this.gdocs.processedAt = maybeDate(data, "gdocs.processed_at");
      this.isProcessing = !this.gdocs.processedAt;
    }

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

  get isGDoc() {
    return this.sourceType === "gdocs";
  }

  get gdocsURL() {
    return !this.isGDoc
      ? ""
      : `https://docs.google.com/document/d/${this.sourceID}/edit`;
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
    if (this.id) {
      return "imported";
    }
    return "draft";
  }
  get statusVerbose() {
    return (
      {
        draft: "Drafting",
        imported: "Imported",
        preview: "Preview",
        embargo: "Embargo",
        released: "Released",
      }[this.status] || "System Error"
    );
  }

  get statusClass() {
    if (this.status === "released") {
      return "is-success";
    }
    if (this.status === "imported") {
      return "is-primary";
    }
    if (this.status === "draft") {
      return "is-danger";
    }
    return "is-warning";
  }

  get adminRoute() {
    return { name: "shared-article-admin", params: { id: this.id } };
  }
  get detailsRoute() {
    return { name: "shared-article", params: { id: this.id } };
  }
  get pageRoute() {
    if (!this.pageID) {
      return null;
    }
    return { name: "news-page", params: { id: this.pageID } };
  }
}

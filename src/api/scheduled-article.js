import Vue from "vue";

import getProp from "@/utils/getter.js";

export default class ScheduledArticle {
  constructor({ id, data, service }) {
    this._url_id = id;
    this._service = service;
    this._reset = JSON.stringify(data);
    this.isSaving = false;
    this.saveError = null;

    this.init(data);

    Vue.observable(this);
  }

  init(data) {
    let props = {
      id: ["ID", ""],
      arcID: ["ArcID", ""],
      body: ["Body", ""],
      blurb: ["Blurb", ""],
      byline: ["Byline", ""],
      hed: ["Hed", ""],
      imageCaption: ["ImageCaption", ""],
      imageCredit: ["ImageCredit", ""],
      imageURL: ["ImageURL", ""],
      linkTitle: ["LinkTitle", ""],
      slug: ["Slug", ""],
      subhead: ["Subhead", ""],
      summary: ["Summary", ""],
      kicker: ["Kicker", ""],
      suppressFeatured: ["SuppressFeatured", false],

      authors: ["Authors", []],
      _scheduleFor: ["ScheduleFor", null],
      _lastArcSync: ["LastArcSync", null],
      _pubDate: ["PubDate", null],
    };

    for (let [key, [val, fallback]] of Object.entries(props)) {
      this[key] = getProp(data, val, { fallback });
    }

    // Date getters
    for (let prop of ["scheduleFor", "lastArcSync", "pubDate"]) {
      let val = this["_" + prop];
      this[prop] = val ? new Date(val) : null;
    }
  }

  reset() {
    this.saveError = null;
    this.init(JSON.parse(this._reset));
  }

  toString() {
    return `Scheduled Article ${this.id}`;
  }

  toJSON() {
    return {
      ID: this.id,
      ArcID: this.arcID,
      Body: this.body,
      Byline: this.byline,
      Blurb: this.blurb,
      Hed: this.hed,
      ImageCaption: this.imageCaption,
      ImageCredit: this.imageCredit,
      ImageURL: this.imageURL,
      LinkTitle: this.linkTitle,
      Slug: this.slug,
      Subhead: this.subhead,
      Summary: this.summary,
      Authors: this.authors,
      ScheduleFor: this.scheduleFor,
      LastArcSync: this.lastArcSync,
      PubDate: this.pubDate,
      Kicker: this.kicker,
      SuppressFeatured: this.suppressFeatured,
    };
  }

  validate() {
    if (!this.kicker) {
      this.saveError = new Error("Kicker must not be blank");
      this.saveError.name = "Validation Error";
      return false;
    }
    if (this.imageURL.match(/^http/)) {
      this.saveError = new Error(
        "Image must be uploaded to Spotlight first (for now)."
      );
      this.saveError.name = "Validation Error";
      return false;
    }
    return true;
  }

  async schedule() {
    if (!this.validate()) {
      return;
    }
    this.isSaving = true;
    this.scheduleFor = this.pubDate;
    this.saveError = await this._service.saveArticle(this._url_id, this);
    this.isSaving = false;
  }
}

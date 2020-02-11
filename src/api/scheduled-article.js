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
    // Date getters
    for (let prop of ["scheduleFor", "lastArcSync", "pubDate"]) {
      Object.defineProperty(this, prop, {
        get() {
          let val = this["_" + prop];
          if (!val) {
            return null;
          }
          return new Date(val);
        },
      });
    }

    Vue.observable(this);
  }

  init(data) {
    let props = {
      id: ["ID", ""],
      arcID: ["ArcID", ""],
      body: ["Body", ""],
      blurb: ["Blurb", ""],
      hed: ["Hed", ""],
      imageCaption: ["ImageCaption", ""],
      imageCredit: ["ImageCredit", ""],
      imageURL: ["ImageURL", ""],
      linkTitle: ["LinkTitle", ""],
      slug: ["Slug", ""],
      subhead: ["Subhead", ""],
      summary: ["Summary", ""],

      authors: ["Authors", []],
      _scheduleFor: ["ScheduleFor", null],
      _lastArcSync: ["LastArcSync", ""],
      _pubDate: ["PubDate", ""],
    };

    for (let [key, [val, fallback]] of Object.entries(props)) {
      this[key] = getProp(data, val, { fallback });
    }

    let [kicker] = getProp(data, "Kicker", { fallback: [""] });
    this.kicker = kicker;
  }

  reset() {
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
      ScheduleFor: this._scheduleFor,
      LastArcSync: this._lastArcSync,
      PubDate: this._pubDate,
      Kicker: [this.kicker],
    };
  }

  async save() {
    this.isSaving = true;
    this.saveError = await this._service.saveArticle(this._url_id, this);
    this.isSaving = false;
  }
}

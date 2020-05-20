import Vue from "vue";

import getProp from "@/utils/getter.js";
import imgproxyURL from "@/api/imgproxy-url.js";

export default class ScheduledArticle {
  constructor({ id, data, client }) {
    this._url_id = id;
    this._client = client;

    this.init(data);

    Vue.observable(this);
  }

  init(data, { save_reset = true } = {}) {
    this.isSaving = false;
    this.saveError = null;
    this._refreshArc = false;

    if (save_reset) {
      this._reset = JSON.stringify(data);
    }

    let props = {
      id: ["InternalID", ""],
      arcID: ["ArcID", ""],
      body: ["Body", ""],
      blurb: ["Blurb", ""],
      byline: ["Byline", ""],
      hed: ["Hed", ""],
      imageDescription: ["ImageDescription", ""],
      imageCaption: ["ImageCaption", ""],
      imageCredit: ["ImageCredit", ""],
      imageURL: ["ImageURL", ""],
      imageSize: ["ImageSize", ""],
      linkTitle: ["LinkTitle", ""],
      slug: ["Slug", ""],
      subhead: ["Subhead", ""],
      summary: ["Summary", ""],
      kicker: ["Kicker", ""],
      suppressFeatured: ["SuppressFeatured", false],
      weight: ["Weight", 0],
      overrideURL: ["OverrideURL", ""],
      aliases: ["Aliases", []],
      modalExclude: ["ModalExclude", false],
      noIndex: ["NoIndex", false],

      warnings: ["Warnings", []],
      authors: ["Authors", []],
      series: ["Series", []],
      topics: ["Topics", []],
    };

    for (let [key, [val, fallback]] of Object.entries(props)) {
      this[key] = getProp(data, val, { fallback });
    }

    // Date getters
    props = {
      scheduleFor: "ScheduleFor",
      lastArcSync: "LastArcSync",
      pubDate: "PubDate",
      lastSaved: "LastSaved",
      lastPublished: "LastPublished",
    };

    let dateObj = {};

    for (let [key, val] of Object.entries(props)) {
      dateObj[key] = getProp(data, val, { fallback: null });
    }

    for (let [prop] of Object.entries(props)) {
      let val = dateObj[prop];
      this[prop] = val ? new Date(val) : null;
    }
  }

  reset() {
    this.init(JSON.parse(this._reset), { save_reset: false });
  }

  toString() {
    return `Scheduled Article ${this.id}`;
  }

  toJSON() {
    return {
      InternalID: this.id,
      ArcID: this.arcID,
      Body: this.body,
      Byline: this.byline,
      Blurb: this.blurb,
      Hed: this.hed,
      ImageCaption: this.imageCaption,
      ImageCredit: this.imageCredit,
      ImageDescription: this.imageDescription,
      ImageURL: this.imageURL,
      LinkTitle: this.linkTitle,
      Slug: this.slug,
      Subhead: this.subhead,
      Summary: this.summary,
      Authors: this.authors,
      Topics: this.topics,
      Series: this.series,
      ScheduleFor: this.scheduleFor,
      LastArcSync: this.lastArcSync,
      PubDate: this.pubDate,
      Kicker: this.kicker,
      SuppressFeatured: this.suppressFeatured,
      ImageSize: this.imageSize,
      OverrideURL: this.overrideURL,
      Aliases: this.aliases,
      ModalExclude: this.modalExclude,
      NoIndex: this.noIndex,
      "almanack-refresh-arc": this._refreshArc,
    };
  }

  deriveSlug() {
    this.slug = this.hed
      .toLowerCase()
      .replace(/\b(the|an?)\b/g, " ")
      .replace(/\bpa\b/g, "pennsylvania")
      .replace(/\W+/g, " ")
      .trim()
      .replace(/ /g, "-");
  }

  get hasPublished() {
    return !!this.lastPublished;
  }

  get pubURL() {
    if (!this.slug) {
      return "";
    }
    let year = this.pubDate.getFullYear();
    let month = (this.pubDate.getMonth() + 1).toString().padStart(2, "0");
    return `https://www.spotlightpa.org/news/${year}/${month}/${this.slug}/`;
  }

  get imagePreviewURL() {
    if (!this.imageURL || this.imageURL.match(/^http/)) {
      return "";
    }
    return imgproxyURL(this.imageURL);
  }

  validate() {
    let valid = true;
    if (!this.kicker) {
      this.saveError = new Error("Kicker must not be blank");
      this.saveError.name = "Validation Error";
      valid = false;
    }
    if (!this.imageURL) {
      this.saveError = new Error("Featured image must not be blank");
      this.saveError.name = "Featured image error";
      valid = false;
    }
    if (!this.slug) {
      this.saveError = new Error("Article slug must not be blank");
      this.saveError.name = "Validation Error";
      valid = false;
    }
    return valid;
  }

  async save({ schedule = null, refreshArc = false }) {
    if (schedule) {
      this.scheduleFor = this.pubDate;
    } else if (schedule !== null) {
      this.scheduleFor = null;
    }
    if (this.scheduleFor && !this.validate() && !refreshArc) {
      return;
    }
    this.isSaving = true;
    this._refreshArc = refreshArc;

    let data;
    [data, this.saveError] = await this._client.saveArticle(this);
    this.isSaving = false;
    this._refreshArc = false;
    if (!this.saveError) {
      this.init(data);
      this.validate();
    }
  }
}

import getProp from "@/utils/getter.js";

export default class ScheduledArticle {
  constructor(data) {
    let props = {
      body: ["Body", ""],
      _pubDate: ["PubDate", ""],
    };

    for (let [key, [val, fallback]] of Object.entries(props)) {
      this[key] = getProp(data, val, { fallback });
    }
  }

  get pubDate() {
    if (!this._pubDate) {
      return null;
    }
    return new Date(this._pubDate);
  }
}

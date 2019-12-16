import { commaAndJoiner } from "@/filters/commaand.js";

import APIArticleImage from "./APIArticleImage.vue";
import APIArticleList from "./APIArticleList.vue";
import APIArticleHTML from "./APIArticleHTML.vue";
import APIArticleHeader from "./APIArticleHeader.vue";
import APIArticleText from "./APIArticleText.vue";
import APIArticlePlaceholder from "./APIArticlePlaceholder.vue";
import APIArticleOEmbed from "./APIArticleOEmbed.vue";

function cmp(a, b) {
  return a === b ? 0 : a < b ? -1 : 1;
}

let contentComponentsTypes = {
  header: APIArticleHeader,
  list: APIArticleList,
  text: APIArticleText,
};

let embedComponentsTypes = {
  code: APIArticleHTML,
  image: APIArticleImage,
  oembed_response: APIArticleOEmbed,
  raw_html: APIArticleHTML,
};

let htmlComponentsTypes = {
  code: block => block.content,
  oembed_response: block => block.raw_oembed.html,
  raw_html: block => block.content,
};

export class Article {
  static from(data) {
    return data.contents
      .map(a => new Article(a))
      .sort((a, b) => cmp(b.plannedDate, a.plannedDate));
  }

  constructor(rawData) {
    let props = {
      actualInchCount: "planning.story_length.inch_count_actual",
      actualLineCount: "planning.story_length.line_count_actual",
      actualWordCount: "planning.story_length.word_count_actual",
      budgetLine: "planning.budget_line",
      description: "description.basic",
      headline: "headlines.basic",
      id: "_id",
      note: "planning.internal_note",
      plannedDate: "planning.scheduling.planned_publish_date",
      plannedWordCount: "planning.story_length.word_count_planned",
      slug: "slug",
      featuredImageCaption: "promo_items.basic.caption",
    };

    this.rawData = rawData;
    for (let [key, val] of Object.entries(props)) {
      this[key] = this.getProp(val);
    }
  }

  getProp(pathStr, { fallback = null } = {}) {
    let obj = this.rawData;
    for (let prop of pathStr.split(".")) {
      if (!obj) {
        break;
      }
      obj = obj[prop];
    }
    if (fallback !== null) {
      return obj || fallback;
    }
    return obj;
  }

  get pubURL() {
    return `https://www.spotlightpa.org${this.rawData.website_url}`;
  }
  get arcURL() {
    return `https://pmn.arcpublishing.com/composer/#!/edit/${this.id}/`;
  }
  get detailsRoute() {
    return { name: "article", params: { id: this.id } };
  }
  get authors() {
    return this.getProp("credits.by", { fallback: [] }).map(item => item.name);
  }
  get byline() {
    return commaAndJoiner(this.authors);
  }
  get status() {
    let statusCode = this.getProp("workflow.status_code");
    return (
      {
        5: "ready",
        6: "published",
      }[statusCode] || "not ready"
    );
  }
  get isPublished() {
    return this.status === "published";
  }
  get featuredImage() {
    let url = this.getProp("promo_items.basic.url", { fallback: "" });
    // Some images haven't been published and can't be used
    if (url.match(/\/public\//)) {
      return url;
    }
    return this.getProp("promo_items.basic.additional_properties.resizeUrl", {
      fallback: "",
    });
  }
  get featuredImageCredits() {
    return this.getProp("promo_items.basic.credits.by", { fallback: [] }).map(
      item => item.byline || item.name
    );
  }

  get contentComponents() {
    let embedcount = 0;

    return this.rawData.content_elements.flatMap(block => {
      let component = contentComponentsTypes[block.type];
      if (component) {
        return {
          component,
          block,
        };
      }
      if (embedComponentsTypes[block.type]) {
        embedcount++;
        let n = embedcount;
        return {
          component: APIArticlePlaceholder,
          block: { n },
        };
      }
      // eslint-disable-next-line no-console
      console.warn("unknown block type", block.type, block);
      return [];
    });
  }

  get embedComponents() {
    let embedcount = 0;

    return this.rawData.content_elements.flatMap(block => {
      let component = embedComponentsTypes[block.type];
      if (!component) {
        return [];
      }
      embedcount++;
      let n = embedcount;
      return {
        component,
        block,
        n,
      };
    });
  }

  get htmlComponents() {
    let embedcount = 0;

    return this.rawData.content_elements.flatMap(block => {
      // Render code blocks but not use placeholder for images
      if (embedComponentsTypes[block.type]) {
        embedcount++;
      }
      let component = contentComponentsTypes[block.type];
      if (component) {
        return {
          component,
          block,
        };
      }
      let renderer = htmlComponentsTypes[block.type];
      if (renderer) {
        return {
          component: {
            render(h) {
              return h("raw-html");
            },
          },
          block: renderer(block),
        };
      }
      if (embedComponentsTypes[block.type]) {
        let n = embedcount;
        return {
          component: APIArticlePlaceholder,
          block: { n },
        };
      }
      // eslint-disable-next-line no-console
      console.warn("unknown block type", block.type, block);
      return [];
    });
  }
}

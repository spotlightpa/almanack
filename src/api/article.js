import getProp from "@/utils/getter.js";
import cmp from "@/utils/cmp.js";
import { commaAndJoiner } from "@/filters/commaand.js";

import APIArticleImage from "@/components/APIArticleImage.vue";
import APIArticleList from "@/components/APIArticleList.vue";
import APIArticleHTML from "@/components/APIArticleHTML.vue";
import APIArticleHeader from "@/components/APIArticleHeader.vue";
import APIArticleText from "@/components/APIArticleText.vue";
import APIArticlePlaceholder from "@/components/APIArticlePlaceholder.vue";
import APIArticleOEmbed from "@/components/APIArticleOEmbed.vue";

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

let ignoreComponentTypes = {
  interstitial_link: true,
};

export default class Article {
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
    return getProp(this.rawData, pathStr, { fallback });
  }

  pubslug() {
    let slug = this.getProp("canonical_url", { fallback: "" });
    let stop = slug.lastIndexOf("-");
    if (stop === -1) {
      return "";
    }
    let start = slug.lastIndexOf("/", stop);
    if (start === -1) {
      return "";
    }
    return slug.slice(start + 1, stop);
  }

  get pubURL() {
    let slug = this.pubslug();
    if (!slug) {
      return "";
    }
    let date = new Date(this.plannedDate);
    let year = date.getFullYear();
    let month = (date.getMonth() + 1).toString().padStart(2, "0");
    return `https://www.spotlightpa.org/news/${year}/${month}/${slug}/`;
  }
  get arcURL() {
    return `https://pmn.arcpublishing.com/composer/edit/${this.id}/`;
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
    let published =
      this.getProp("additional_properties.is_published") ||
      this.getProp("additional_properties.has_published_copy");
    if (published) {
      return "published";
    }
    let statusCode = this.getProp("workflow.status_code");
    return (
      {
        1: "notReadyWorking",
        2: "notReadyAssigning",
        3: "notReadySecondEdit",
        4: "notReadyRim",
        5: "readySlot",
        6: "readyDone",
      }[statusCode] || "unknown"
    );
  }
  get statusVerbose() {
    return (
      {
        notReadyWorking: "Working",
        notReadyAssigning: "Assigning",
        notReadySecondEdit: "Second Edit",
        notReadyRim: "Rim",
        readySlot: "Slot",
        readyDone: "Done",
        published: "Published",
      }[this.status] || "Unknown"
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
      if (ignoreComponentTypes[block.type]) {
        return [];
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
      if (ignoreComponentTypes[block.type]) {
        return [];
      }
      // eslint-disable-next-line no-console
      console.warn("unknown block type", block.type, block);
      return [];
    });
  }
}

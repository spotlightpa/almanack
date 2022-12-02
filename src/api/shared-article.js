// TODO: Use embargo dates for stuff
import { h, reactive } from "vue";

import { intcomma } from "journalize";

import getProp from "@/utils/getter.js";
import cmp from "@/utils/cmp.js";
import commaAnd from "@/utils/comma-and.js";

import { formatDate } from "@/utils/time-format.js";

import ArcArticleDivider from "@/components/ArcArticleDivider.vue";
import ArcArticleImage from "@/components/ArcArticleImage.vue";
import ArcArticleList from "@/components/ArcArticleList.vue";
import ArcArticleHTML from "@/components/ArcArticleHTML.vue";
import ArcArticleHeader from "@/components/ArcArticleHeader.vue";
import ArcArticleText from "@/components/ArcArticleText.vue";
import ArcArticlePlaceholder from "@/components/ArcArticlePlaceholder.vue";
import ArcArticleOEmbed from "@/components/ArcArticleOEmbed.vue";

let contentComponentsTypes = {
  header: ArcArticleHeader,
  list: ArcArticleList,
  text: ArcArticleText,
  divider: ArcArticleDivider,
};

let embedComponentsTypes = {
  code: ArcArticleHTML,
  image: ArcArticleImage,
  oembed_response: ArcArticleOEmbed,
  raw_html: ArcArticleHTML,
};

let htmlComponentsTypes = {
  code: (block) => block.content,
  oembed_response: (block) => block.raw_oembed.html,
  raw_html: (block) => block.content,
};

let ignoreComponentTypes = {
  interstitial_link: true,
};

export default class SharedArticle {
  static from(data) {
    return data.contents
      .map((a) => reactive(new SharedArticle(a)))
      .sort((a, b) => cmp(b.plannedDate, a.plannedDate));
  }

  constructor(rawData) {
    this.init(rawData);
  }

  init(rawData) {
    let props = {
      actualInchCount: "raw_data.planning.story_length.inch_count_actual",
      actualLineCount: "raw_data.planning.story_length.line_count_actual",
      actualWordCount: "raw_data.planning.story_length.word_count_actual",
      budgetLine: "raw_data.planning.budget_line",
      description: "raw_data.description.basic",
      headline: "raw_data.headlines.basic",
      id: "source_id",
      note: "note",
      plannedDate: "raw_data.planning.scheduling.planned_publish_date",
      plannedWordCount: "raw_data.planning.story_length.word_count_planned",
      slug: "raw_data.slug",
      featuredImageCaption: "raw_data.promo_items.basic.caption",
    };

    this.rawData = rawData;
    for (let [key, val] of Object.entries(props)) {
      this[key] = this.getProp(val);
    }
    this._almanackStatus = this.getProp("status", { fallback: 0 });
  }

  getProp(pathStr, { fallback = null } = {}) {
    return getProp(this.rawData, pathStr, { fallback });
  }

  pubslug() {
    let slug = this.getProp("raw_data.canonical_url", { fallback: "" });
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

  get isPlanned() {
    return this._almanackStatus === "P" || this._almanackStatus === "S";
  }

  get isAvailable() {
    return this._almanackStatus === "S";
  }

  get arcURL() {
    return `https://pmn.arcpublishing.com/composer/edit/${this.id}/`;
  }
  get detailsRoute() {
    return { name: "article", params: { id: this.id } };
  }
  get spotlightPARedirectRoute() {
    return { name: "redirect-arc-news-page", params: { id: this.id } };
  }
  get authors() {
    return this.getProp("raw_data.credits.by", { fallback: [] }).map((item) => {
      let byline = getProp(
        item,
        "raw_data.additional_properties.original.byline"
      );
      if (byline) {
        return byline;
      }
      let { name } = item;
      // Hack for bad names with orgs in them
      if (/ of /.test(name)) {
        return name;
      }
      let org = getProp(item, "org");
      if (org) {
        return `${name} of ${org}`;
      }
      return name;
    });
  }
  get byline() {
    return commaAnd(this.authors);
  }
  get status() {
    if (this.isAvailable) {
      return "published";
    }
    let published =
      this.getProp("raw_data.additional_properties.is_published") ||
      this.getProp("raw_data.additional_properties.has_published_copy");
    if (published) {
      return "readyPublished";
    }
    let statusCode = this.getProp("raw_data.workflow.status_code");
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
        readyPublished: "Released",
        published: "Ready",
      }[this.status] || "Unknown"
    );
  }
  get featuredImage() {
    let srcURL = this.getProp("raw_data.promo_items.basic.url", {
      fallback: "",
    });
    // Some images haven't been published and can't be used
    if (!srcURL.match(/\/public\//)) {
      srcURL = this.getProp(
        "raw_data.promo_items.basic.additional_properties.resizeUrl",
        {
          fallback: "",
        }
      );
    }
    if (!srcURL) {
      return "";
    }
    return `/api/proxy-image/${window.btoa(srcURL)}`;
  }
  get featuredImageCredits() {
    return this.getProp("raw_data.promo_items.basic.credits.by", {
      fallback: [],
    }).map((item) => item.name || item.byline);
  }

  get contentComponents() {
    let embedcount = 0;

    return this.rawData.content_elements.flatMap((block) => {
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
          component: ArcArticlePlaceholder,
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

    return this.rawData.content_elements.flatMap((block) => {
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

    return this.rawData.content_elements.flatMap((block) => {
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
            render() {
              return h("raw-html");
            },
          },
          block: renderer(block),
        };
      }
      if (embedComponentsTypes[block.type]) {
        let n = embedcount;
        return {
          component: ArcArticlePlaceholder,
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

  get emailSubject() {
    return `New Spotlight PA story ${this.slug}`;
  }

  get emailBody() {
    let text = `
New ${this.slug}

https://almanack.data.spotlightpa.org/articles/${this.id}

Planned for ${formatDate(this.plannedDate)}${
      this.note ? `\n\nPublication Notes:\n\n${this.note}` : ""
    }

Budget:

${this.budgetLine}

Word count planned: ${intcomma(this.plannedWordCount)}
Word count actual: ${intcomma(this.actualWordCount)}
Lines: ${this.actualLineCount}
Column inches: ${this.actualInchCount}
`;
    return text.trim();
  }
}

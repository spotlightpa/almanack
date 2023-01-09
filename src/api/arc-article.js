import { h } from "vue";

import getProp from "@/utils/getter.js";
import commaAnd from "@/utils/comma-and.js";
import maybeDate from "@/utils/maybe-date.js";

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

export default class ArcArticle {
  constructor(rawData) {
    let props = {
      actualInchCount: "planning.story_length.inch_count_actual",
      actualLineCount: "planning.story_length.line_count_actual",
      actualWordCount: "planning.story_length.word_count_actual",
      budgetLine: "planning.budget_line",
      description: "description.basic",
      headline: "headlines.basic",
      id: "_id",
      note: "almanack-note",
      plannedWordCount: "planning.story_length.word_count_planned",
      slug: "slug",
      featuredImageCaption: "promo_items.basic.caption",
    };

    this.rawData = rawData;
    for (let [key, val] of Object.entries(props)) {
      this[key] = this.getProp(val);
    }
    this.plannedDate = maybeDate(
      rawData,
      "planning.scheduling.planned_publish_date"
    );
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

  get arcURL() {
    return `https://pmn.arcpublishing.com/composer/edit/${this.id}/`;
  }
  get authors() {
    return this.getProp("credits.by", { fallback: [] }).map((item) => {
      let byline = getProp(item, "additional_properties.original.byline");
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
  get featuredImage() {
    let srcURL = this.getProp("promo_items.basic.url", { fallback: "" });
    // Some images haven't been published and can't be used
    if (!srcURL.match(/\/public\//)) {
      srcURL = this.getProp(
        "promo_items.basic.additional_properties.resizeUrl",
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
    return this.getProp("promo_items.basic.credits.by", { fallback: [] }).map(
      (item) => item.name || item.byline
    );
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
}

<script>
import { commaAndJoiner } from "@/filters/commaand.js";

import APIArticleContentImage from "./APIArticleContentImage.vue";
import APIArticleContentList from "./APIArticleContentList.vue";
import APIArticleContentHTML from "./APIArticleContentHTML.vue";
import APIArticleContentHeader from "./APIArticleContentHeader.vue";
import APIArticleContentText from "./APIArticleContentText.vue";
import APIArticleContentPlaceholder from "./APIArticleContentPlaceholder.vue";
import APIArticleContentOEmbed from "./APIArticleContentOEmbed.vue";

let contentComponentsTypes = {
  text: APIArticleContentText,
  header: APIArticleContentHeader,
  list: APIArticleContentList
};

let embedComponentsTypes = {
  image: APIArticleContentImage,
  code: APIArticleContentHTML,
  raw_html: APIArticleContentHTML,
  oembed_response: APIArticleContentOEmbed
};

let htmlComponentsTypes = {
  code: block => block.content,
  raw_html: block => block.content,
  oembed_response: block => block.raw_oembed.html
};

class Article {
  static from(rawData) {
    const getter = pathStr =>
      pathStr
        .split(".")
        .reduce((xs, x) => (xs && xs[x] ? xs[x] : null), rawData);

    let props = {
      note: "planning.internal_note",
      budgetLine: "planning.budget_line",
      description: "description.basic",
      headline: "headlines.basic",
      plannedDate: "planning.scheduling.planned_publish_date",
      plannedWordCount: "planning.story_length.word_count_planned",
      actualWordCount: "planning.story_length.word_count_actual",
      actualLineCount: "planning.story_length.line_count_actual",
      actualInchCount: "planning.story_length.inch_count_actual",
      slug: "slug",
      id: "_id"
    };

    let article = new Article();
    article.rawData = rawData;
    for (let [key, val] of Object.entries(props)) {
      article[key] = getter(val);
    }
    return article;
  }

  get pubURL() {
    return "https://www.spotlightpa.org" + this.rawData.website_url;
  }
  get authors() {
    return this.rawData.credits.by.map(item => item.name);
  }
  get byline() {
    return commaAndJoiner(this.authors);
  }
  get status() {
    return (
      {
        5: "ready",
        6: "published"
      }[this.rawData.workflow.status_code] || "not ready"
    );
  }
  get isPublished() {
    return this.status === "published";
  }

  get contentComponents() {
    let embedcount = 0;

    return this.rawData.content_elements.flatMap(block => {
      let component = contentComponentsTypes[block.type];
      if (component) {
        return {
          component,
          block
        };
      }
      if (embedComponentsTypes[block.type]) {
        embedcount++;
        let n = embedcount;
        return {
          component: APIArticleContentPlaceholder,
          block: { n }
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
        n
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
          block
        };
      }
      let renderer = htmlComponentsTypes[block.type];
      if (renderer) {
        return {
          component: {
            render(h) {
              return h("raw-html");
            }
          },
          block: renderer(block)
        };
      }
      if (embedComponentsTypes[block.type]) {
        let n = embedcount;
        return {
          component: APIArticleContentPlaceholder,
          block: { n }
        };
      }
      // eslint-disable-next-line no-console
      console.warn("unknown block type", block.type, block);
      return [];
    });
  }
}

export default {
  name: "APIArticle",
  props: {
    data: { type: Object, required: true }
  },
  data() {
    return { article: Article.from(this.data) };
  }
};
</script>

<template>
  <div>
    <slot :article="article"></slot>
  </div>
</template>

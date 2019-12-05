<script>
import { commaAndJoiner } from "@/filters/commaand.js";

// import APIArticleContentImage from "./APIArticleContentImage.vue";
import APIArticleContentText from "./APIArticleContentText.vue";
import APIArticleContentPlaceholder from "./APIArticleContentPlaceholder.vue";

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

    return this.rawData.content_elements.flatMap(block =>
      elementToComponent(block)
    );

    function elementToComponent(block) {
      let component = {
        text: APIArticleContentText
      }[block.type];
      if (!component) {
        embedcount++;
        let n = embedcount;
        return {
          component: APIArticleContentPlaceholder,
          block: { n }
        };
      }
      return {
        component,
        block
      };
    }
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

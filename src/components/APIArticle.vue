<script>
export default {
  name: "APIArticle",
  props: {
    data: { type: Object, required: true }
  },
  data() {
    return { article: {} };
  },
  created() {
    const getter = pathStr =>
      pathStr
        .split(".")
        .reduce((xs, x) => (xs && xs[x] ? xs[x] : null), this.data);

    let props = {
      budgetLine: "planning.budget_line",
      description: "description.basic",
      headline: "headlines.basic",
      plannedDate: "planning.scheduling.planned_publish_date",
      plannedWordCount: "planning.story_length.word_count_planned",
      actualWordCount: "planning.story_length.word_count_actual",
      actualLineCount: "planning.story_length.line_count_actual",
      actualInchCount: "planning.story_length.inch_count_actual",
      slug: "slug"
    };
    for (let [key, val] of Object.entries(props)) {
      this.article[key] = getter(val);
    }
    this.article.isPublished = this.data.workflow.status_code === 6;
    this.article.pubURL = "https://www.spotlightpa.org" + this.data.website_url;
    this.article.rawData = this.data;
  }
};
</script>

<template>
  <div>
    <slot :article="article"></slot>
  </div>
</template>

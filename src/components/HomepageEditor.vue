<script>
export default {
  props: {
    pages: Array,
    editorsPicks: Object,
  },

  methods: {
    push(article) {
      this.editorsPicks.featuredStories.push(article);
    },
  },
};
</script>

<template>
  <div class="columns">
    <div class="column is-half">
      <PageSelector :pages="pages" @select="push($event)" />
    </div>
    <div class="column is-half">
      <BulmaField label="Homepage featured article">
        <HomepageEditorDraggable v-model="editorsPicks.featuredStories" />
        Pin top story on homepage
      </BulmaField>
      <BulmaField label="Subfeatures stories">
        <HomepageEditorDraggable v-model="editorsPicks.subfeatures" />

        Bulleted items under the top story
      </BulmaField>
      <BulmaField label="Pinned stories">
        <HomepageEditorDraggable v-model="editorsPicks.topSlots" />
        Pin stories at the top of homepage
      </BulmaField>
      <BulmaField label="Homepage topper">
        <HomepageEditorDraggable v-model="editorsPicks.topper" />
        If populated, a topper will appear on the homepage with the first story
        as its featured item and the other stories as bulleted items underneath
      </BulmaField>
    </div>
  </div>
</template>

<style scoped>
.overflow {
  text-overflow: ellipsis;
  overflow-x: hidden;
  display: block;
}
.select-none {
  cursor: grab;
  user-select: none;
}
</style>

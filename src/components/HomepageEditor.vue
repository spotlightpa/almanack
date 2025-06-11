<script>
/* eslint-disable vue/no-mutating-props */
export default {
  props: {
    editorsPicks: Object,
    showCallout: Boolean,
    showInvestigation: Boolean,
    showImpact: Boolean,
  },

  methods: {
    push(article) {
      this.editorsPicks.featuredStories.push(article.filePath);
    },
  },
};
</script>

<template>
  <div class="columns">
    <div class="column is-half">
      <PageFinder class="is-sticky" @select-page="push($event)"></PageFinder>
    </div>
    <div class="column is-half">
      <BulmaField label="Homepage featured article">
        <HomepageEditorDraggable
          v-model="editorsPicks.featuredStories"
        ></HomepageEditorDraggable>
        Pin top story on homepage
      </BulmaField>
      <BulmaField label="Subfeatures stories">
        <HomepageEditorDraggable
          v-model="editorsPicks.subfeatures"
        ></HomepageEditorDraggable>

        Bulleted items under the top story
      </BulmaField>
      <BulmaField label="Pinned stories">
        <HomepageEditorDraggable
          v-model="editorsPicks.topSlots"
        ></HomepageEditorDraggable>
        Pin stories at the top left of homepage
      </BulmaField>
      <BulmaField v-if="showCallout" label="Editor's Picks Homepage Callout">
        <HomepageEditorDraggable
          v-model="editorsPicks.edCallout"
        ></HomepageEditorDraggable>
        Requires at least two stories
      </BulmaField>
      <BulmaField v-if="showInvestigation" label="Featured Investigations">
        <HomepageEditorDraggable
          v-model="editorsPicks.edInvestigations"
        ></HomepageEditorDraggable>
        Shows an investigation on a black background
      </BulmaField>
      <BulmaField v-if="showImpact" label="Featured Impact">
        <HomepageEditorDraggable
          v-model="editorsPicks.edImpact"
        ></HomepageEditorDraggable>
        Requires at least two stories
      </BulmaField>
    </div>
  </div>
</template>

<style scoped>
.is-sticky {
  position: sticky;
  top: 0px;
}
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

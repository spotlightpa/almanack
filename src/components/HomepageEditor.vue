<script>
/* eslint-disable vue/no-mutating-props */
export default {
  props: {
    editorsPicks: Object,
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
        Pin stories at the top of homepage
      </BulmaField>
      <BulmaField label="Homepage topper">
        <HomepageEditorDraggable
          v-model="editorsPicks.topper"
        ></HomepageEditorDraggable>
        If populated, a topper will appear on the homepage with the first story
        as its featured item and the other stories as bulleted items underneath
      </BulmaField>
      <BulmaField
        v-if="editorsPicks.edCallout"
        label="Editor's Picks Homepage Callout"
      >
        <HomepageEditorDraggable
          v-model="editorsPicks.edCallout"
        ></HomepageEditorDraggable>
        New section for homepage redesign
      </BulmaField>
      <BulmaField
        v-if="editorsPicks.edInvestigations"
        label="Featured Investigations"
      >
        <HomepageEditorDraggable
          v-model="editorsPicks.edInvestigations"
        ></HomepageEditorDraggable>
        New section for homepage redesign
      </BulmaField>
      <BulmaField v-if="editorsPicks.edImpact" label="Featured Impact">
        <HomepageEditorDraggable
          v-model="editorsPicks.edImpact"
        ></HomepageEditorDraggable>
        New section for homepage redesign
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

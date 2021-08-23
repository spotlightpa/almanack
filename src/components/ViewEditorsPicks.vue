<script>
import Vue from "vue";
import { reactive, computed, toRefs } from "@vue/composition-api";

import { useClient, makeState } from "@/api/hooks.js";

import EditorsPicks from "./EditorsPicks.vue";

class EditorsPicksData {
  constructor(data, articlesByPath) {
    this._initialData = { data, articlesByPath };
    this.reset();
    Vue.observable(this);
  }

  reset() {
    let { data, articlesByPath } = this._initialData;
    for (let prop of [
      "featuredStories",
      "subfeatures",
      "topSlots",
      "sidebarPicks",
    ]) {
      let a = data?.[prop] ?? [];
      this[prop] = a.map((s) => articlesByPath.get(s)).filter((a) => a);
    }

    this.limitSubfeatures = data?.limitSubfeatures ?? false;
    this.subfeaturesLimit = data?.subfeaturesLimit ?? 0;
  }

  toJSON() {
    const getPath = (a) => a.file_path;
    return {
      featuredStories: this.featuredStories.map(getPath),
      subfeatures: this.subfeatures.map(getPath),
      topSlots: this.topSlots.map(getPath),
      sidebarPicks: this.sidebarPicks.map(getPath),
      limitSubfeatures: !!this.limitSubfeatures,
      subfeaturesLimit: +this.subfeaturesLimit,
    };
  }
}

const toArticle = (a) => ({
  ...a,
  filterableProps: `${a.internal_id} ${a.hed} ${a.authors.join(" ")}`,
});

export default {
  name: "ViewEditorsPicks",
  components: {
    EditorsPicks,
  },
  metaInfo: {
    title: "Homepage Editor",
  },
  setup() {
    let { listAllPages, getEditorsPicks, saveEditorsPicks } = useClient();

    let { apiState: listState, exec: listExec } = makeState();
    let { apiState: edPicksState, exec: edPickExec } = makeState();

    let state = reactive({
      didLoad: computed(() => listState.didLoad && edPicksState.didLoad),
      isLoading: computed(() => listState.isLoading || edPicksState.isLoading),
      error: computed(() => listState.error ?? edPicksState.error),
      pages: computed(() =>
        listState.rawData ? listState.rawData.pages.map(toArticle) : []
      ),
      articlesByPath: computed(
        () => new Map(state.pages.map((a) => [a.file_path, a]))
      ),
      edPicks: computed(
        () =>
          state.didLoad &&
          new EditorsPicksData(edPicksState.rawData, state.articlesByPath)
      ),
    });

    let actions = {
      reload() {
        return Promise.all([
          listExec(listAllPages),
          edPickExec(getEditorsPicks),
        ]);
      },
      save() {
        return edPickExec(() => saveEditorsPicks(state.edPicks));
      },
    };

    actions.reload();

    return {
      ...toRefs(state),
      ...actions,
    };
  },
};
</script>

<template>
  <div>
    <nav class="breadcrumb has-succeeds-separator" aria-label="breadcrumbs">
      <ul>
        <li>
          <router-link :to="{ name: 'admin' }">Admin</router-link>
        </li>
        <li class="is-active">
          <router-link exact :to="{ name: 'editors-picks' }">
            Homepage Editor
          </router-link>
        </li>
      </ul>
    </nav>

    <h1 class="title">Homepage Editor</h1>

    <progress
      v-if="!didLoad && isLoading"
      class="progress is-large is-warning"
      max="100"
    >
      Loading…
    </progress>

    <EditorsPicks
      v-if="didLoad && edPicks"
      :articles="pages"
      :editors-picks="edPicks"
    />

    <div v-if="error" class="message is-danger">
      <div class="message-header">{{ error.name }}</div>
      <div class="message-body">
        <p class="content">{{ error.message }}</p>
        <div class="buttons">
          <button
            class="button is-danger has-text-weight-semibold"
            @click="reload"
          >
            Reload?
          </button>
        </div>
      </div>
    </div>

    <div class="buttons">
      <button
        type="button"
        class="button is-primary has-text-weight-semibold"
        :disabled="isLoading"
        :class="{ 'is-loading': isLoading }"
        @click="save"
      >
        Save
      </button>
      <button
        type="button"
        class="button is-light has-text-weight-semibold"
        :disabled="isLoading"
        :class="{ 'is-loading': isLoading }"
        @click="edPicks.reset()"
      >
        Revert
      </button>
    </div>

    <details class="content">
      <summary>Help</summary>

      <p>The Spotlight PA homepage is defined by three lists:</p>
      <ul>
        <li>
          Homepage featured article: Defines what is in the first featured image
          on the page.
        </li>
        <li>Subfeatures stories: Bulleted items under the featured story.</li>
        <li>
          Pinned stories: Other stories that will appear on the homepage,
          regardless of date.
        </li>
      </ul>
      <p>
        Search for stories by slug, title, or author in the recent Spotlight PA
        articles box. Stories must be saved in Almanack to be listed. Click a
        story to add it to the featured list or drag it to a particular list.
      </p>
      <p>
        All lists filter out <em>unpublished</em> stories automatically.
        Duplicated items are removed automatically. For the featured slot and
        restricted subfeatures, the last published items in the list win.
      </p>
      <p>
        Example: Suppose SPL1, SPL2, SPL3, and SPL4 are all in the featured
        article list and the subfeatures list, but SPL4 is not published yet.
        Suppose also that the subfeatures list is limited to two items. In that
        case, SPL3 will be in the top slot and SPL1 and SPL2 will be listed as
        bulleted items. When SPL4 is published, it will take the featured slot,
        and SPL2 and SPL3 will become bulleted items. In this way, you can
        arrange the homepage so that it automatically changes as stories become
        published.
      </p>
    </details>
  </div>
</template>

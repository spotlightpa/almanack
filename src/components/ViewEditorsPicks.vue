<script>
import Vue from "vue";
import { reactive, computed, toRefs } from "@vue/composition-api";

import { useClient } from "@/api/hooks.js";

import EditorsPicks from "./EditorsPicks.vue";

class EditorsPicksData {
  constructor(data, articlesByPath) {
    this._initialData = { data, articlesByPath };
    this.reset();
    Vue.observable(this);
  }

  reset() {
    let { data, articlesByPath } = this._initialData;
    for (let prop of ["featuredStories", "subfeatures", "topSlots"]) {
      let a = data?.[prop] ?? [];
      this[prop] = a.map((s) => articlesByPath.get(s)).filter((a) => a);
    }

    this.limitSubfeatures = data?.limitSubfeatures ?? false;
    this.subfeaturesLimit = data?.subfeaturesLimit ?? 0;
  }

  toJSON() {
    const getPath = (a) => a.spotlightpa_path;
    return {
      featuredStories: this.featuredStories.map(getPath),
      subfeatures: this.subfeatures.map(getPath),
      topSlots: this.topSlots.map(getPath),
      limitSubfeatures: !!this.limitSubfeatures,
      subfeaturesLimit: +this.subfeaturesLimit,
    };
  }
}

export default {
  name: "ViewEditorsPicks",
  components: {
    EditorsPicks,
  },
  metaInfo: {
    title: "Homepage Editor",
  },
  setup() {
    let {
      listSpotlightPAArticles,
      getEditorsPicks,
      saveEditorsPicks,
    } = useClient();

    const toArticle = (a) => ({
      ...a,
      filterableProps: `${a.internal_id} ${a.hed} ${a.authors.join(" ")}`,
    });

    let listState = reactive({
      isLoading: false,
      data: { articles: [] },
      error: null,
    });

    let edPicksState = reactive({
      isLoading: false,
      data: null,
      error: null,
    });

    let state = reactive({
      didLoad: false,
      isLoading: computed(() => listState.isLoading || edPicksState.isLoading),
      error: computed(() => listState.error ?? edPicksState.error),
      articles: computed(() => listState.data.articles.map(toArticle)),
      articlesByPath: computed(
        () => new Map(state.articles.map((a) => [a.spotlightpa_path, a]))
      ),
      edPicks: computed(
        () => new EditorsPicksData(edPicksState.data, state.articlesByPath)
      ),
    });

    let actions = {
      async fetch(stateObj, cb) {
        stateObj.isLoading = true;
        let data;
        [data, stateObj.error] = await cb();
        stateObj.isLoading = false;
        if (stateObj.error) {
          return false;
        }
        stateObj.data = data;
        return true;
      },
      async reload() {
        let results = await Promise.all([
          actions.fetch(listState, () => listSpotlightPAArticles()),
          actions.fetch(edPicksState, () => getEditorsPicks()),
        ]);
        state.didLoad = state.didLoad || results.every((r) => !!r);
      },
      save() {
        return actions.fetch(edPicksState, () =>
          saveEditorsPicks(state.edPicks)
        );
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

    <h1 class="title">
      Homepage Editor
    </h1>

    <details class="content">
      <summary>Help</summary>

      <p>Lorem ipsum</p>
    </details>

    <progress
      v-if="!didLoad && isLoading"
      class="progress is-large is-warning"
      max="100"
    >
      Loading…
    </progress>

    <EditorsPicks
      v-if="didLoad"
      :articles="articles"
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
    <details>
      <summary>debug</summary>
      <code class="code">{{ edPicks }}</code>
    </details>
  </div>
</template>

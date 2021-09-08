<script>
import Vue from "vue";
import { reactive, computed, toRefs, watch } from "@vue/composition-api";

import { useClient, makeState } from "@/api/hooks.js";
import { formatDateTime } from "@/utils/time-format.js";

import BulmaField from "./BulmaField.vue";
import EditorsPicks from "./EditorsPicks.vue";

class Page {
  constructor(data) {
    this.id = data.id;
    this.filePath = data.file_path ?? "";
    this.internalID = data.internal_id ?? "";
    this.hed = data.hed ?? "";
    this.authors = data.authors ?? [];
    this.filterableProps = `${this.internalID} ${this.hed} ${this.authors.join(
      " "
    )}`;
  }
}

class EditorsPicksData {
  constructor(siteConfig, pagesByPath) {
    this._initialData = { siteConfig, pagesByPath };
    this.reset();
    Vue.observable(this);
  }

  reset() {
    let { siteConfig, pagesByPath } = this._initialData;
    for (let prop of [
      "featuredStories",
      "subfeatures",
      "topSlots",
      "sidebarPicks",
    ]) {
      let a = siteConfig.data?.[prop] ?? [];
      this[prop] = a.map((s) => pagesByPath.get(s)).filter((a) => !!a);
    }
    this.id = siteConfig.id;
    this.scheduleFor = siteConfig.schedule_for;
    let pub = siteConfig.published_at;
    this.publishedAt = pub.Valid ? new Date(pub.Time) : null;
    this.isCurrent = !!this.publishedAt;
  }

  clone(scheduleFor) {
    let { siteConfig, pagesByPath } = this._initialData;
    let newPick = new EditorsPicksData(
      {
        ...siteConfig,
        id: 0,
        published_at: { Valid: false },
        schedule_for: scheduleFor,
      },
      pagesByPath
    );
    return newPick;
  }

  toJSON() {
    const getPath = (a) => a.filePath;
    return {
      id: this.id,
      schedule_for: this.scheduleFor,
      data: {
        featuredStories: this.featuredStories.map(getPath),
        subfeatures: this.subfeatures.map(getPath),
        topSlots: this.topSlots.map(getPath),
        sidebarPicks: this.sidebarPicks.map(getPath),
      },
    };
  }
}

export default {
  name: "ViewEditorsPicks",
  components: {
    BulmaField,
    EditorsPicks,
  },
  metaInfo: {
    title: "Homepage Editor",
  },
  setup() {
    let removedItems = [];
    let { listAllPages, getEditorsPicks, saveEditorsPicks } = useClient();

    let { apiState: listState, exec: listExec } = makeState();
    let { apiState: edPicksState, exec: edPickExec } = makeState();

    let state = reactive({
      didLoad: computed(() => listState.didLoad && edPicksState.didLoad),
      isLoading: computed(() => listState.isLoading || edPicksState.isLoading),
      error: computed(() => listState.error ?? edPicksState.error),
      pages: computed(() =>
        listState.rawData ? listState.rawData.pages.map((p) => new Page(p)) : []
      ),
      pagesByPath: computed(
        () => new Map(state.pages.map((p) => [p.filePath, p]))
      ),
      pagesAndPicks: computed(() => ({
        pages: state.pages,
        rawPicks: edPicksState.rawData?.datums ?? [],
      })),
      allEdPicks: [],
      nextSchedule: null,
    });

    let actions = {
      reload() {
        return Promise.all([
          listExec(listAllPages),
          edPickExec(getEditorsPicks),
        ]);
      },
      save() {
        return edPickExec(() =>
          saveEditorsPicks({
            datums: state.allEdPicks,
            removed_items: removedItems,
          })
        );
      },
    };
    watch(
      () => state.pagesAndPicks,
      ({ pages, rawPicks }) => {
        if (!pages.length || !rawPicks.length) {
          return;
        }
        removedItems = [];
        state.allEdPicks = rawPicks.map(
          (data) => new EditorsPicksData(data, state.pagesByPath)
        );
      }
    );

    actions.reload();

    return {
      ...toRefs(state),
      ...actions,

      formatDateTime,

      async addScheduledPicks() {
        let lastPick = state.allEdPicks[state.allEdPicks.length - 1];
        state.allEdPicks.push(lastPick.clone(state.nextSchedule));
        state.nextSchedule = null;
        await this.$nextTick();
        // TODO: Fix this array if we ever upgrade to Vue 3
        // https://vueuse.org/core/useTemplateRefsList/
        this.$refs.edPicksEls[this.$refs.edPicksEls.length - 1].scrollIntoView({
          behavior: "smooth",
          block: "start",
        });
      },
      removeScheduledPick(i) {
        let id = state.allEdPicks[i].id;
        if (id) {
          removedItems.push(id);
        }
        state.allEdPicks.splice(i, 1);
      },
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

    <template v-if="allEdPicks.length">
      <div v-for="(edpick, i) of allEdPicks" :key="i" class="p-4 zebra-row">
        <h2 ref="edPicksEls" class="title">
          {{
            edpick.isCurrent
              ? "Current Homepage"
              : `Scheduled for ${formatDateTime(edpick.scheduleFor)}`
          }}
        </h2>
        <EditorsPicks :pages="pages" :editors-picks="edpick" />
        <button
          v-if="!edpick.isCurrent"
          type="button"
          class="button is-danger has-text-weight-semibold"
          @click="removeScheduledPick(i)"
        >
          Remove
        </button>
      </div>
      <h2 class="mt-2 title">Add a scheduled change</h2>
      <BulmaField v-slot="{ idForLabel }" label="Schedule for">
        <b-datetimepicker
          :id="idForLabel"
          v-model="nextSchedule"
          icon="user-clock"
          :datetime-formatter="formatDateTime"
          :inline="true"
          locale="en-US"
        />
        <button
          type="button"
          :disabled="!nextSchedule || nextSchedule < new Date()"
          class="mt-3 button is-success has-text-weight-semibold"
          @click="addScheduledPicks"
        >
          <span class="icon is-size-6">
            <font-awesome-icon :icon="['fas', 'plus']" />
          </span>
          <span>Add</span>
        </button>
      </BulmaField>
    </template>
    <progress
      v-if="!didLoad && isLoading"
      class="progress is-large is-warning"
      max="100"
    >
      Loading…
    </progress>

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

<style scoped>
.zebra-row {
  background-color: #fff;
}

.zebra-row:nth-child(even) {
  background-color: #fafafa;
}

.zebra-row + .zebra-row {
  border-top: 1px solid #dbdbdb;
}
</style>

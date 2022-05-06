<script>
import Vue from "vue";
import { reactive, computed, toRefs, watch } from "@vue/composition-api";

import { useClient, makeState } from "@/api/hooks.js";
import { formatDateTime } from "@/utils/time-format.js";

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
    this.pagesByPath = pagesByPath;
    this.reset(siteConfig);
    Vue.observable(this);
  }

  reset(siteConfig) {
    for (let prop of [
      "featuredStories",
      "subfeatures",
      "topSlots",
      "sidebarPicks",
      "topper",
    ]) {
      let a = siteConfig.data?.[prop] ?? [];
      this[prop] = a.map((s) => this.pagesByPath.get(s)).filter((a) => !!a);
    }
    this.scheduleFor = siteConfig.schedule_for;
    let pub = siteConfig.published_at;
    this.publishedAt = pub ? new Date(pub) : null;
    this.isCurrent = !!this.publishedAt;
  }

  clone(scheduleFor) {
    let { data } = JSON.parse(JSON.stringify(this));
    let newPick = new EditorsPicksData(
      {
        schedule_for: scheduleFor,
        data,
        published_at: null,
      },
      this.pagesByPath
    );
    return newPick;
  }

  toJSON() {
    const getPath = (a) => a.filePath;
    return {
      schedule_for: this.scheduleFor,
      data: {
        featuredStories: this.featuredStories.map(getPath),
        subfeatures: this.subfeatures.map(getPath),
        topSlots: this.topSlots.map(getPath),
        sidebarPicks: this.sidebarPicks.map(getPath),
        topper: this.topper.map(getPath),
      },
    };
  }
}

export default {
  name: "ViewEditorsPicks",
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
      pages: computed(
        () => listState.rawData?.pages.map((p) => new Page(p)) ?? []
      ),
      pagesByPath: computed(
        () => new Map(state.pages.map((p) => [p.filePath, p]))
      ),
      rawPicks: computed(() => edPicksState.rawData?.configs ?? []),

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
          })
        );
      },
      reset() {
        let { pages, rawPicks } = state;

        if (!pages.length || !rawPicks.length) {
          return;
        }

        state.allEdPicks = rawPicks.map(
          (data) => new EditorsPicksData(data, state.pagesByPath)
        );
      },
    };
    watch(
      () => [state.pages, state.rawPicks],
      () => actions.reset()
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
        state.allEdPicks.splice(i, 1);
      },
    };
  },
};
</script>

<template>
  <div>
    <BulmaBreadcrumbs
      :links="[
        { name: 'Admin', to: { name: 'admin' } },
        { name: 'Homepage Editor', to: { name: 'editors-picks' } },
      ]"
    />

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
          Remove {{ formatDateTime(edpick.scheduleFor) }}
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
        @click="reset"
      >
        Revert
      </button>
    </div>

    <SpinnerProgress :is-loading="isLoading" />
    <ErrorReloader :error="error" @reload="reload" />
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

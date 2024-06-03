<script>
import { reactive, computed, toRefs, watch } from "vue";

import { useClient } from "@/api/client.js";
import { makeState } from "@/api/service-util.js";
import Page from "@/api/spotlightpa-all-pages-item.js";

import { formatDateTime, today, tomorrow } from "@/utils/time-format.js";
import useScrollTo from "@/utils/use-scroll-to.js";
import maybeDate from "@/utils/maybe-date.js";

class EditorsPicksData {
  constructor(siteConfig, pagesByPath) {
    this.pagesByPath = pagesByPath;
    this.reset(siteConfig);
  }

  reset(siteConfig) {
    for (let prop of ["featuredStories", "subfeatures", "topSlots", "topper"]) {
      let a = siteConfig.data?.[prop] ?? [];
      this[prop] = a.map((s) => this.pagesByPath.get(s)).filter((a) => !!a);
    }
    this.scheduleFor = maybeDate(siteConfig, "schedule_for");
    this.publishedAt = maybeDate(siteConfig, "published_at");
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
        topper: this.topper.map(getPath),
      },
    };
  }
}

export default {
  setup() {
    const [container, scrollTo] = useScrollTo();

    let { listAllPages, getStateCollegeEditor, saveStateCollegeEditor } =
      useClient();
    let { apiState: listState, exec: listExec } = makeState();
    let { apiState: edPicksState, exec: edPickExec } = makeState();
    let state = reactive({
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
          edPickExec(getStateCollegeEditor),
        ]);
      },
      save() {
        return edPickExec(() =>
          saveStateCollegeEditor({
            configs: state.allEdPicks,
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
      () => actions.reset(),
      { deep: true }
    );
    actions.reload();
    return {
      container,

      ...toRefs(state),
      ...actions,
      formatDateTime,
      async addScheduledPicks() {
        let lastPick = state.allEdPicks[state.allEdPicks.length - 1];
        state.allEdPicks.push(lastPick.clone(state.nextSchedule));
        state.nextSchedule = null;
        await scrollTo();
      },
      removeScheduledPick(i) {
        state.allEdPicks.splice(i, 1);
      },
      today,
      tomorrow,
    };
  },
};
</script>

<template>
  <MetaHead>
    <title>State College Frontpage Editor • Spotlight PA Almanack</title>
  </MetaHead>
  <div>
    <BulmaBreadcrumbs
      :links="[
        { name: 'Admin', to: { name: 'admin' } },
        {
          name: 'State College Frontpage Editor',
          to: { name: 'state-college-editor' },
        },
      ]"
    ></BulmaBreadcrumbs>

    <h1 class="title">State College Frontpage Editor</h1>

    <div v-if="allEdPicks.length" ref="container">
      <div v-for="(edpick, i) of allEdPicks" :key="i" class="p-4 zebra-row">
        <h2 data-scroll-to class="title">
          {{
            edpick.isCurrent
              ? "Current Frontpage"
              : `Scheduled for ${formatDateTime(edpick.scheduleFor)}`
          }}
        </h2>
        <HomepageEditor :editors-picks="edpick"></HomepageEditor>
        <button
          v-if="!edpick.isCurrent"
          type="button"
          class="button is-danger has-text-weight-semibold"
          @click="removeScheduledPick(i)"
        >
          Remove {{ formatDateTime(edpick.scheduleFor) }}
        </button>
      </div>
      <h2 class="mt-2 mb-0 title is-size-3">Add a scheduled change</h2>
      <BulmaDateTime
        v-model="nextSchedule"
        label="Schedule for"
        icon="user-clock"
      >
        <p class="mt-2 buttons">
          <button
            type="button"
            :disabled="!nextSchedule || nextSchedule < new Date() || null"
            class="button is-small is-success has-text-weight-semibold"
            @click="addScheduledPicks"
          >
            <span class="icon is-size-6">
              <font-awesome-icon :icon="['fas', 'plus']"></font-awesome-icon>
            </span>
            <span>Add</span>
          </button>
          <button
            class="button is-small is-light has-text-weight-semibold"
            type="button"
            @click="nextSchedule = today()"
          >
            Today
          </button>
          <button
            type="button"
            class="button is-small is-light has-text-weight-semibold"
            @click="nextSchedule = tomorrow()"
          >
            Tomorrow
          </button>
        </p>
      </BulmaDateTime>
    </div>

    <div class="mt-5 buttons">
      <button
        type="button"
        class="button is-primary has-text-weight-semibold"
        :disabled="isLoading || null"
        :class="{ 'is-loading': isLoading }"
        @click="save"
      >
        Save
      </button>
      <button
        type="button"
        class="button is-light has-text-weight-semibold"
        :disabled="isLoading || null"
        :class="{ 'is-loading': isLoading }"
        @click="reset"
      >
        Revert
      </button>
    </div>

    <SpinnerProgress :is-loading="isLoading"></SpinnerProgress>
    <ErrorReloader :error="error" @reload="reload"></ErrorReloader>
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

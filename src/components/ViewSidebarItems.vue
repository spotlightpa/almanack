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

class SidebarData {
  constructor(siteConfig, pagesByPath) {
    this.pagesByPath = pagesByPath;
    this.reset(siteConfig);
    Vue.observable(this);
  }

  reset(siteConfig) {
    this.items = siteConfig?.data.items ?? [];
    this.scheduleFor = siteConfig.schedule_for;
    let pub = siteConfig.published_at;
    this.publishedAt = pub ? new Date(pub) : null;
    this.isCurrent = !!this.publishedAt;
  }

  clone(scheduleFor) {
    let { data } = JSON.parse(JSON.stringify(this));
    let newPicks = new SidebarData(
      {
        schedule_for: scheduleFor,
        data,
        published_at: null,
      },
      this.pagesByPath
    );
    return newPicks;
  }

  toJSON() {
    return {
      schedule_for: this.scheduleFor,
      data: {
        items: this.items,
      },
    };
  }
}

export default {
  name: "ViewSidebarItems",
  metaInfo: {
    title: "Sidebar Editor",
  },
  setup() {
    let { listAllPages, getSidebar, saveSidebar } = useClient();

    let { apiState: listState, exec: listExec } = makeState();
    let { apiState: sidebarState, exec: sidebarExec } = makeState();

    let state = reactive({
      didLoad: computed(() => listState.didLoad && sidebarState.didLoad),
      isLoading: computed(() => listState.isLoading || sidebarState.isLoading),
      error: computed(() => listState.error ?? sidebarState.error),
      pages: computed(
        () => listState.rawData?.pages.map((p) => new Page(p)) ?? []
      ),
      pagesByPath: computed(
        () => new Map(state.pages.map((p) => [p.filePath, p]))
      ),
      rawSidebars: computed(() => sidebarState.rawData?.configs ?? []),

      allSidebars: [],
      nextSchedule: null,
    });

    let actions = {
      reload() {
        return Promise.all([listExec(listAllPages), sidebarExec(getSidebar)]);
      },
      save() {
        return sidebarExec(() =>
          saveSidebar({
            configs: state.allSidebars,
          })
        );
      },
      reset() {
        let { pages, rawSidebars } = state;

        if (!pages.length || !rawSidebars.length) {
          return;
        }

        state.allSidebars = rawSidebars.map(
          (data) => new SidebarData(data, state.pagesByPath)
        );
      },
    };
    watch(
      () => [state.pages, state.rawSidebars],
      () => actions.reset()
    );

    actions.reload();

    return {
      ...toRefs(state),
      ...actions,

      formatDateTime,

      async addScheduledPicks() {
        let lastPick =
          state.allSidebars[state.allSidebars.length - 1] ??
          new SidebarData({ data: { items: [] } }, state.pages);
        state.allSidebars.push(lastPick.clone(state.nextSchedule));
        state.nextSchedule = null;
        await this.$nextTick();
        // TODO: Fix this array if we ever upgrade to Vue 3
        // https://vueuse.org/core/useTemplateRefsList/
        this.$refs.sidebarEls[this.$refs.sidebarEls.length - 1].scrollIntoView({
          behavior: "smooth",
          block: "start",
        });
      },
      removeScheduledPick(i) {
        state.allSidebars.splice(i, 1);
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
        { name: 'Sidebar Items', to: { name: 'sidebar-items' } },
      ]"
    />

    <h1 class="title">Sidebar Items</h1>

    <template v-if="allSidebars.length">
      <div v-for="(sidebar, i) of allSidebars" :key="i" class="p-4 zebra-row">
        <h2 ref="sidebarEls" class="title">
          {{
            sidebar.isCurrent
              ? "Current Sidebar"
              : `Scheduled for ${formatDateTime(sidebar.scheduleFor)}`
          }}
        </h2>
        <div class="columns">
          <div class="column is-half">
            <PageSelector :pages="pages" />
          </div>
        </div>
        <button
          v-if="!sidebar.isCurrent"
          type="button"
          class="button is-danger has-text-weight-semibold"
          @click="removeScheduledPick(i)"
        >
          Remove {{ formatDateTime(sidebar.scheduleFor) }}
        </button>
      </div>
    </template>
    <template v-if="!isLoading">
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

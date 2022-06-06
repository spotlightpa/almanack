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

let itemIds = 0;

class SidebarData {
  constructor(siteConfig) {
    this.reset(siteConfig);
    Vue.observable(this);
  }

  reset(siteConfig) {
    let items = siteConfig?.data.items ?? [];
    this.items = items.map((item) => ({
      item,
      id: itemIds++,
    }));
    this.scheduleFor = siteConfig.schedule_for;
    let pub = siteConfig.published_at;
    this.publishedAt = pub ? new Date(pub) : null;
    this.isCurrent = !!this.publishedAt;
  }

  add({ filePath }) {
    let item = {
      label: "Editor’s Pick",
      labelColor: "#ff6c36",
      linkColor: "#000000",
      backgroundColor: "#f5f5f5",
      page: filePath,
    };
    this.items.push({
      item,
      id: itemIds++,
    });
  }

  swap({ pos, dir }) {
    let target = pos + dir;
    if (target < 0 || target >= this.items.length) {
      // eslint-disable-next-line no-console
      console.error("bad swap", pos, target);
      return;
    }
    let firstIndex = Math.min(pos, target);
    let first = this.items[firstIndex];
    let last = this.items[Math.max(pos, target)];
    // Must use splice for Vue reactivity
    this.items.splice(firstIndex, 2, last, first);
  }

  remove(pos) {
    this.items.splice(pos, 1);
  }

  clone(scheduleFor) {
    let { data } = JSON.parse(JSON.stringify(this));
    let newPicks = new SidebarData({
      schedule_for: scheduleFor,
      data,
      published_at: null,
    });
    return newPicks;
  }

  toJSON() {
    return {
      schedule_for: this.scheduleFor,
      data: {
        items: this.items.map((obj) => obj.item),
      },
    };
  }
}

export default {
  metaInfo: {
    title: "Sidebar Editor",
  },
  setup() {
    let { listAllPages, getSidebar, saveSidebar } = useClient();
    let { apiState: pagesState, exec: pagesExec } = makeState();
    let { apiState: sidebarState, exec: sidebarExec } = makeState();
    let state = reactive({
      pages: computed(
        () => pagesState.rawData?.pages.map((p) => new Page(p)) ?? []
      ),
      pagesByPath: computed(
        () => new Map(state.pages.map((p) => [p.filePath, p]))
      ),
      rawSidebars: computed(() => sidebarState.rawData?.configs ?? []),
      allSidebars: [],
      nextSchedule: null,
    });
    let actions = {
      reloadSidebars() {
        return sidebarExec(getSidebar);
      },
      reloadPages() {
        return pagesExec(listAllPages);
      },
      save() {
        return sidebarExec(() =>
          saveSidebar({
            configs: state.allSidebars,
          })
        );
      },
      initSidebars() {
        let { rawSidebars } = state;
        if (!rawSidebars.length) {
          return;
        }
        state.allSidebars = rawSidebars.map((data) => new SidebarData(data));
      },
    };
    watch(
      () => [state.rawSidebars],
      () => actions.initSidebars()
    );
    actions.reloadSidebars();
    actions.reloadPages();
    return {
      sidebarState,
      pagesState,
      ...toRefs(state),
      ...actions,
      formatDateTime,
      async addScheduledPicks() {
        let lastPick = state.allSidebars[state.allSidebars.length - 1];
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
        <div class="columns is-multiline">
          <div class="column is-full">
            <SidebarItem
              v-for="({ item, id }, pos) of sidebar.items"
              :key="id"
              class="p-4 zebra-row"
              :item="item"
              :pos="pos"
              :length="sidebar.items.length"
              @swap="sidebar.swap($event)"
              @remove="sidebar.remove($event)"
            />
          </div>
          <div class="column is-full">
            <h2 class="title">Add new item</h2>
            <PageSelector :pages="pages" @select="sidebar.add($event)" />
            <SpinnerProgress :is-loading="pagesState.isLoading" />
            <ErrorReloader :error="pagesState.error" @reload="reloadPages" />
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
    <template v-if="!sidebarState.isLoading">
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
        :disabled="sidebarState.isLoading"
        :class="{ 'is-loading': sidebarState.isLoading }"
        @click="save"
      >
        Save
      </button>
      <button
        type="button"
        class="button is-light has-text-weight-semibold"
        :disabled="sidebarState.isLoading"
        :class="{ 'is-loading': sidebarState.isLoading }"
        @click="reloadSidebars"
      >
        Revert
      </button>
    </div>

    <SpinnerProgress :is-loading="sidebarState.isLoading" />
    <ErrorReloader :error="sidebarState.error" @reload="reloadSidebars" />
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

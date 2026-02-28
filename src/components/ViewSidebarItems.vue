<script setup>
import { reactive, computed, ref, watch } from "vue";

import { get, post, getSidebar, saveSidebar } from "@/api/client-v2.js";
import { makeState } from "@/api/service-util.js";
import { formatDateTime, today, tomorrow } from "@/utils/time-format.js";
import useScrollTo from "@/utils/use-scroll-to.js";
import maybeDate from "@/utils/maybe-date.js";

let itemIds = 0;

class SidebarData {
  constructor(siteConfig) {
    this.reset(siteConfig);
  }

  reset(siteConfig) {
    let items = siteConfig?.data.items ?? [];
    this.items = items.map((item) => ({
      item: { ...item },
      id: itemIds++,
    }));
    this.scheduleFor = maybeDate(siteConfig, "schedule_for");
    this.publishedAt = maybeDate(siteConfig, "published_at");
    this.isCurrent = !!this.publishedAt;
  }

  add({ filePath }) {
    let item = {
      label: "Editor's Pick",
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
    let newPicks = reactive(
      new SidebarData({
        schedule_for: scheduleFor,
        data,
        published_at: null,
      })
    );
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

const [container, scrollTo] = useScrollTo();

const { apiState: sidebarState, exec: sidebarExec } = makeState();

const rawSidebars = computed(() => sidebarState.rawData?.configs ?? []);
const allSidebars = ref([]);
const nextSchedule = ref(null);

function reloadSidebars() {
  return sidebarExec(() => get(getSidebar));
}

function save() {
  return sidebarExec(() =>
    post(saveSidebar, {
      configs: allSidebars.value,
    })
  );
}

function initSidebars() {
  if (!rawSidebars.value.length) {
    return;
  }
  allSidebars.value = rawSidebars.value.map((data) =>
    reactive(new SidebarData(data))
  );
}

watch(
  () => [rawSidebars.value],
  () => initSidebars(),
  { deep: true }
);

reloadSidebars();

async function addScheduledPicks() {
  let lastPick = allSidebars.value[allSidebars.value.length - 1];
  allSidebars.value.push(lastPick.clone(nextSchedule.value));
  nextSchedule.value = null;
  await scrollTo();
}

function removeScheduledPick(i) {
  allSidebars.value.splice(i, 1);
}
</script>

<template>
  <MetaHead>
    <title>Sidebar Editor • Spotlight PA Almanack</title>
  </MetaHead>
  <div>
    <BulmaBreadcrumbs
      :links="[
        { name: 'Admin', to: { name: 'admin' } },
        { name: 'Sidebar Items', to: { name: 'sidebar-items' } },
      ]"
    ></BulmaBreadcrumbs>

    <h1 class="title">Sidebar Items</h1>

    <div v-if="allSidebars.length" ref="container">
      <div v-for="(sidebar, i) of allSidebars" :key="i" class="p-4 zebra-row">
        <h2 data-scroll-to class="title">
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
            ></SidebarItem>
          </div>
          <div class="column is-full">
            <h2 class="mb-1 title is-size-3">Add new item</h2>
            <PageFinder @select-page="sidebar.add($event)"></PageFinder>
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
    </div>
    <template v-if="!sidebarState.isLoading">
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
    </template>

    <div class="mt-5 buttons">
      <button
        type="button"
        class="button is-primary has-text-weight-semibold"
        :disabled="sidebarState.isLoading || null"
        :class="{ 'is-loading': sidebarState.isLoading }"
        @click="save"
      >
        Save
      </button>
      <button
        type="button"
        class="button is-light has-text-weight-semibold"
        :disabled="sidebarState.isLoading || null"
        :class="{ 'is-loading': sidebarState.isLoading }"
        @click="reloadSidebars"
      >
        Revert
      </button>
    </div>

    <SpinnerProgress :is-loading="sidebarState.isLoading"></SpinnerProgress>
    <ErrorReloader
      :error="sidebarState.error"
      @reload="reloadSidebars"
    ></ErrorReloader>
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

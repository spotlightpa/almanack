<script setup>
import { ref, watch } from "vue";

import { get, post, getSiteParams, postSiteParams } from "@/api/client-v2.js";
import { makeState } from "@/api/service-util.js";
import { useFileList } from "@/api/file-list.js";

import { formatDateTime, today, tomorrow } from "@/utils/time-format.js";
import useScrollTo from "@/utils/use-scroll-to.js";
import maybeDate from "@/utils/maybe-date.js";

class SiteParamsModel {
  constructor(config) {
    this.scheduleFor = maybeDate(config, "schedule_for");
    this.publishedAt = maybeDate(config, "published_at");
    this.isCurrent = !!this.publishedAt;
    this.data = config.data;
  }

  toJSON() {
    return {
      schedule_for: this.scheduleFor,
      data: this.data,
    };
  }
}

const scheduledConfigs = ref([]);
const nextSchedule = ref(null);

const { exec, apiStateRefs } = makeState();

function fetch() {
  return exec(() => get(getSiteParams));
}

const [container, scrollTo] = useScrollTo();

async function addScheduledConfig() {
  let lastParams = scheduledConfigs.value[scheduledConfigs.value.length - 1];
  scheduledConfigs.value.push(
    new SiteParamsModel({
      ...JSON.parse(JSON.stringify(lastParams)),
      schedule_for: nextSchedule.value,
    })
  );
  nextSchedule.value = null;
  await scrollTo();
}

function removeScheduledConfig(i) {
  scheduledConfigs.value.splice(i, 1);
}

async function save() {
  let configs = scheduledConfigs.value;
  scheduledConfigs.value = [];
  return exec(() => post(postSiteParams, { configs }));
}

watch(apiStateRefs.rawData, (data) => {
  if (!data.configs) {
    return;
  }
  scheduledConfigs.value = data.configs.map((cfg) => new SiteParamsModel(cfg));
});

const { isLoading, isLoadingThrottled, error } = apiStateRefs;

const files = useFileList();

fetch();
</script>

<template>
  <MetaHead>
    <title>Sitewide Settings • Spotlight PA Almanack</title>
  </MetaHead>
  <div>
    <div class="px-2">
      <BulmaBreadcrumbs
        :links="[
          { name: 'Admin', to: { name: 'admin' } },
          { name: 'Sitewide Settings', to: { name: 'site-params' } },
        ]"
      />
      <h1 class="title">Sitewide Settings</h1>
    </div>
    <div v-if="scheduledConfigs.length" ref="container">
      <div
        v-for="(params, i) of scheduledConfigs"
        :key="i"
        class="px-2 py-4 zebra-row"
      >
        <h2 data-scroll-to class="title is-3">
          {{
            params.isCurrent
              ? "Current Settings"
              : `Scheduled for ${formatDateTime(params.scheduleFor)}`
          }}
        </h2>

        <SiteParams :params="params" :file-props="files" />

        <button
          v-if="!params.isCurrent"
          type="button"
          class="mt-2 button is-danger has-text-weight-semibold"
          @click="removeScheduledConfig(i)"
        >
          Remove {{ formatDateTime(params.scheduleFor) }}
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
            @click="addScheduledConfig"
          >
            <span class="icon is-size-6">
              <font-awesome-icon :icon="['fas', 'plus']" />
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
        :class="{ 'is-loading': isLoadingThrottled }"
        @click="save"
      >
        Save
      </button>
      <button
        type="button"
        class="button is-light has-text-weight-semibold"
        :disabled="isLoading || null"
        :class="{ 'is-loading': isLoadingThrottled }"
        @click="fetch"
      >
        Revert
      </button>
    </div>

    <SpinnerProgress :is-loading="isLoadingThrottled" />
    <ErrorReloader :error="error" @reload="fetch" />
  </div>
</template>

<style scoped>
.zebra-row {
  background-color: #fff;
}

.zebra-row:nth-child(odd) {
  background-color: #fafafa;
}

.zebra-row + .zebra-row {
  border-top: 1px solid #dbdbdb;
}
</style>

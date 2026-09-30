<script setup lang="ts">
import { ref, type Ref, watch } from "vue";

import { formatDateTime, today, tomorrow } from "@/utils/time-format.ts";
import { get, post, getSiteData, postSiteData } from "@/api/client.ts";
import { makeState } from "@/api/loader.ts";
import { useFileList } from "@/api/file-list.ts";
import useScrollTo from "@/utils/use-scroll-to.ts";
import maybeDate from "@/utils/maybe-date.ts";

const props = defineProps({
  title: { type: String, required: true },
  breadcrumbTo: { type: [String, Object], default: "" },
  location: { type: String, required: true },
});

class SiteParamsModel {
  scheduleFor: Date | null;
  publishedAt: Date | null;
  isCurrent: boolean;
  data: Record<string, unknown>;

  constructor(config: Record<string, unknown>) {
    this.scheduleFor = maybeDate(config, "schedule_for");
    this.publishedAt = maybeDate(config, "published_at");
    this.isCurrent = !!this.publishedAt;
    this.data = (config.data ?? {}) as Record<string, unknown>;
  }

  toJSON() {
    return {
      schedule_for: this.scheduleFor,
      data: this.data,
    };
  }
}

const query = `?location=${props.location}`;

const scheduledConfigs = ref<SiteParamsModel[]>([]);
const siteParamsComps = ref<{ saveParams(): unknown }[]>([]);
const nextSchedule = ref<Date | null>(null);

const { exec, apiStateRefs } = makeState();

function fetch() {
  return exec(() => get(getSiteData + query));
}

const [container, scrollTo] = useScrollTo();

async function addScheduledConfig() {
  let lastParams = scheduledConfigs.value[
    scheduledConfigs.value.length - 1
  ] ?? { data: {} };
  scheduledConfigs.value.push(
    new SiteParamsModel({
      ...JSON.parse(JSON.stringify(lastParams)),
      schedule_for: nextSchedule.value,
    })
  );
  nextSchedule.value = null;
  await scrollTo();
}

function removeScheduledConfig(i: number) {
  scheduledConfigs.value.splice(i, 1);
}

async function save() {
  let configs = siteParamsComps.value.map((comp) => comp.saveParams());
  return exec(() => post(postSiteData + query, { configs }));
}

watch(apiStateRefs.rawData as Ref<Record<string, unknown>>, (data) => {
  if (!data?.configs) {
    return;
  }
  scheduledConfigs.value = (data.configs as Record<string, unknown>[]).map(
    (cfg) => new SiteParamsModel(cfg)
  );
});

const { isLoading, isLoadingThrottled, error } = apiStateRefs;

const files = useFileList();

fetch();
</script>

<template>
  <div>
    <div class="px-2">
      <BulmaBreadcrumbs
        :links="[
          { name: 'Admin', to: { name: 'admin' } },
          { name: title, to: breadcrumbTo },
        ]"
      ></BulmaBreadcrumbs>
      <h1 class="title">{{ title }}</h1>
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

        <slot
          name="form"
          :params="params"
          :file-props="files"
          :set-ref="
            (el: { saveParams(): unknown } | null) => {
              if (el) siteParamsComps[i] = el;
            }
          "
        ></slot>

        <button
          v-if="!params.isCurrent"
          type="button"
          class="mt-2 button is-danger has-text-weight-semibold"
          @click="removeScheduledConfig(i)"
        >
          Remove {{ formatDateTime(params.scheduleFor) }}
        </button>
      </div>
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
          :disabled="!nextSchedule || nextSchedule < new Date() || undefined"
          class="button is-small is-success has-text-weight-semibold"
          @click="addScheduledConfig"
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
    <div class="mt-5 buttons">
      <button
        type="button"
        class="button is-primary has-text-weight-semibold"
        :disabled="isLoading || undefined"
        :class="{ 'is-loading': isLoadingThrottled }"
        @click="save"
      >
        Save
      </button>
      <button
        type="button"
        class="button is-light has-text-weight-semibold"
        :disabled="isLoading || undefined"
        :class="{ 'is-loading': isLoadingThrottled }"
        @click="fetch"
      >
        Revert
      </button>
    </div>

    <SpinnerProgress :is-loading="isLoadingThrottled"></SpinnerProgress>
    <ErrorReloader :error="error" @reload="fetch"></ErrorReloader>
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

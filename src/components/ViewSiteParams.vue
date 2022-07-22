<script>
import Vue, { ref, nextTick, reactive, toRefs, watch } from "vue";

import { useClient, makeState } from "@/api/hooks.js";
import { useFileList } from "@/api/file-list.js";

import { formatDateTime } from "@/utils/time-format.js";

class SiteParams {
  constructor(config) {
    let scheduleFor = config.schedule_for;
    this.scheduleFor = scheduleFor ? new Date(scheduleFor) : null;
    let pub = config.published_at;
    this.publishedAt = pub ? new Date(pub) : null;
    this.isCurrent = !!this.publishedAt;
    this.data = config.data;
    Vue.observable(this);
  }

  toJSON() {
    return {
      schedule_for: this.scheduleFor,
      data: this.data,
    };
  }
}

export default {
  metaInfo: {
    title: "Sitewide Settings",
  },
  setup() {
    const container = ref();

    let { getSiteParams, postSiteParams } = useClient();
    const { apiState, exec } = makeState();

    const state = reactive({
      ...toRefs(apiState),

      configs: [],
      nextSchedule: null,
    });

    let actions = {
      fetch() {
        state.configs = [];
        return exec(() => getSiteParams());
      },
      save() {
        let { configs } = state;
        state.configs = [];
        return exec(() => postSiteParams({ configs }));
      },
      init() {
        if (!apiState.rawData) {
          return;
        }
        state.configs = apiState.rawData.configs.map(
          (data) => new SiteParams(data)
        );
      },
      async addScheduledConfig() {
        let lastParams = state.configs[state.configs.length - 1];
        state.configs.push(
          new SiteParams({
            ...JSON.parse(JSON.stringify(lastParams)),
            schedule_for: state.nextSchedule,
          })
        );
        state.nextSchedule = null;
        await nextTick();

        let el = container.value;
        let headings = el.querySelectorAll("[data-heading]");
        let newPick = Array.from(headings).at(-1);
        newPick.scrollIntoView({
          behavior: "smooth",
          block: "start",
        });
      },
      removeScheduledConfig(i) {
        state.configs.splice(i, 1);
      },
    };

    watch(
      () => apiState.rawData,
      () => actions.init()
    );

    actions.fetch();

    return {
      container,
      ...toRefs(state),
      ...actions,

      formatDateTime,
      files: useFileList(),
    };
  },
};
</script>

<template>
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

    <div v-if="configs.length" ref="container">
      <div v-for="(params, i) of configs" :key="i" class="px-2 py-4 zebra-row">
        <h2 data-heading class="title is-3">
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
      <h2 class="mt-2 title">Create a scheduled change</h2>
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
          @click="addScheduledConfig"
        >
          <span class="icon is-size-6">
            <font-awesome-icon :icon="['fas', 'plus']" />
          </span>
          <span>Add</span>
        </button>
      </BulmaField>
    </div>

    <div class="mt-1 buttons">
      <button
        type="button"
        class="button is-primary has-text-weight-semibold"
        :disabled="isLoading"
        :class="{ 'is-loading': isLoadingThrottled }"
        @click="save"
      >
        Save
      </button>
      <button
        type="button"
        class="button is-light has-text-weight-semibold"
        :disabled="isLoading"
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

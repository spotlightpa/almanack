<script setup>
import { formatDateTime, today, tomorrow } from "@/utils/time-format.ts";

const props = defineProps({
  title: { type: String, required: true },
  breadcrumbTo: { type: [String, Object], default: "" },
  editor: { type: Object, required: true },
});

const {
  scheduledConfigs,
  siteParamsComps,
  nextSchedule,
  container,
  isLoading,
  isLoadingThrottled,
  error,
  files,
  fetch,
  save,
  addScheduledConfig,
  removeScheduledConfig,
} = props.editor;
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
            (el) => {
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
          :disabled="!nextSchedule || nextSchedule < new Date() || null"
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

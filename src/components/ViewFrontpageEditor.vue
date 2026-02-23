<script>
import { get, post, getSiteData, postSiteData } from "@/api/client-v2.js";
import usePicks from "@/api/editors-picks.js";
import { useRoute } from "vue-router";

import { formatDateTime, today, tomorrow } from "@/utils/time-format.js";
import useScrollTo from "@/utils/use-scroll-to.js";

export default {
  setup() {
    const route = useRoute();
    const { dataFile, title, showCallout, showInvestigation, showImpact } =
      route.meta;

    const [container, scrollTo] = useScrollTo();
    const picks = usePicks({
      fetchData: () => get(getSiteData + "?location=" + dataFile),
      saveData: (data) => post(postSiteData + "?location=" + dataFile, data),
    });

    return {
      container,
      title,
      showCallout,
      showInvestigation,
      showImpact,
      ...picks,
      formatDateTime,
      async addScheduledPicks() {
        let lastPick =
          picks.allEdPicks.value[picks.allEdPicks.value.length - 1];
        picks.allEdPicks.value.push(lastPick.clone(picks.nextSchedule.value));
        picks.nextSchedule.value = null;
        await scrollTo();
      },
      removeScheduledPick(i) {
        picks.allEdPicks.value.splice(i, 1);
      },
      today,
      tomorrow,
    };
  },
};
</script>

<template>
  <MetaHead>
    <title>{{ title }} • Spotlight PA Almanack</title>
  </MetaHead>
  <div>
    <BulmaBreadcrumbs
      :links="[
        { name: 'Admin', to: { name: 'admin' } },
        { name: title, to: { name: $route.name } },
      ]"
    ></BulmaBreadcrumbs>

    <h1 class="title">{{ title }}</h1>

    <div v-if="allEdPicks.length" ref="container">
      <div v-for="(edpick, i) of allEdPicks" :key="i" class="p-4 zebra-row">
        <h2 data-scroll-to class="title">
          {{
            edpick.isCurrent
              ? "Current layout"
              : `Scheduled for ${formatDateTime(edpick.scheduleFor)}`
          }}
        </h2>
        <HomepageEditor
          :editors-picks="edpick"
          :show-callout="showCallout"
          :show-investigation="showInvestigation"
          :show-impact="showImpact"
        ></HomepageEditor>
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
        :disabled="isLoadingThrottled || null"
        :class="{ 'is-loading': isLoadingThrottled }"
        @click="save"
      >
        Save
      </button>
      <button
        type="button"
        class="button is-light has-text-weight-semibold"
        :disabled="isLoadingThrottled || null"
        :class="{ 'is-loading': isLoadingThrottled }"
        @click="reset"
      >
        Revert
      </button>
    </div>
    <SpinnerProgress :is-loading="isLoadingThrottled"></SpinnerProgress>
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

import { reactive, computed, toRefs, watch } from "vue";

import { makeState } from "@/api/service-util.js";
import maybeDate from "@/utils/maybe-date.js";

class EditorsPicksData {
  constructor(siteConfig) {
    this.reset(siteConfig);
  }

  reset(siteConfig) {
    for (let prop of [
      "featuredStories",
      "subfeatures",
      "topSlots",
      "edImpact",
      "edInvestigations",
      "edCallout",
    ]) {
      let a = siteConfig.data?.[prop] ?? [];
      this[prop] = [...a];
    }
    this.scheduleFor = maybeDate(siteConfig, "schedule_for");
    this.publishedAt = maybeDate(siteConfig, "published_at");
    this.isCurrent = !!this.publishedAt;
  }

  clone(scheduleFor) {
    let { data } = JSON.parse(JSON.stringify(this));
    let newPick = new EditorsPicksData({
      schedule_for: scheduleFor,
      data,
      published_at: null,
    });
    return newPick;
  }

  toJSON() {
    return {
      schedule_for: this.scheduleFor,
      data: {
        featuredStories: this.featuredStories,
        subfeatures: this.subfeatures,
        topSlots: this.topSlots,
        edImpact: this.edImpact,
        edInvestigations: this.edInvestigations,
        edCallout: this.edCallout,
      },
    };
  }
}

export default function usePicks({ fetchData, saveData }) {
  let { apiStateRefs: edPicksState, exec: edPickExec } = makeState();
  let state = reactive({
    rawPicks: computed(() => edPicksState.rawData.value?.configs ?? []),
    allEdPicks: [],
    nextSchedule: null,
  });
  let actions = {
    reload() {
      return edPickExec(fetchData);
    },
    save() {
      return edPickExec(() =>
        saveData({
          configs: state.allEdPicks,
        })
      );
    },
    reset() {
      let { rawPicks } = state;
      if (!rawPicks.length) {
        return;
      }
      state.allEdPicks = rawPicks.map((data) => new EditorsPicksData(data));
    },
  };
  watch(
    () => state.rawPicks,
    () => actions.reset(),
    { deep: true }
  );
  actions.reload();

  return {
    isLoadingThrottled: edPicksState.isLoadingThrottled,
    error: edPicksState.error,

    ...toRefs(state),
    ...actions,
  };
}

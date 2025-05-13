import { reactive, computed, toRefs, watch } from "vue";

import {
  get,
  post,
  getStateCollegeEditor,
  saveStateCollegeEditor,
} from "@/api/client-v2.js";
import { makeState } from "@/api/service-util.js";

import maybeDate from "@/utils/maybe-date.js";

class EditorsPicksData {
  constructor(siteConfig) {
    this.reset(siteConfig);
  }

  reset(siteConfig) {
    for (let prop of ["featuredStories", "subfeatures", "topSlots", "topper"]) {
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
        topper: this.topper,
      },
    };
  }
}

export default function usePicks() {
  let { apiStateRefs: edPicksState, exec: edPickExec } = makeState();
  let state = reactive({
    rawPicks: computed(() => edPicksState.rawData.value?.configs ?? []),
    allEdPicks: [],
    nextSchedule: null,
  });
  let actions = {
    reload() {
      return edPickExec(() => get(getStateCollegeEditor));
    },
    save() {
      return edPickExec(() =>
        post(saveStateCollegeEditor, {
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

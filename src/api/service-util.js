import { computed, reactive, toRefs, watch } from "vue";

import { useThrottleToggle } from "@/utils/throttle.js";

export function makeState() {
  const apiState = reactive({
    rawData: null,
    isLoading: false,
    error: null,
    isLoadingThrottled: null,
  });

  let apiStateRefs = toRefs(apiState);
  apiState.isLoadingThrottled = useThrottleToggle(apiStateRefs.isLoading);

  return {
    apiState,
    apiStateRefs,

    async exec(callback) {
      if (apiState.isLoading) {
        return;
      }
      apiState.isLoading = true;
      let data;
      [data, apiState.error] = await callback();
      apiState.isLoading = false;
      if (!apiState.error) {
        apiState.rawData = data;
      }
    },
  };
}

export function watchAPI(watchCb, fetcher) {
  const { exec, apiStateRefs } = makeState();
  const doFetch = (newVal) => exec(() => fetcher(newVal));

  watch(watchCb, doFetch, { immediate: true });

  return {
    apiState: apiStateRefs,
    async fetch() {
      return doFetch(watchCb());
    },
    computer(cb) {
      return computed(() => cb(apiStateRefs.rawData.value));
    },
    computedObj(cb) {
      return computed(() => {
        let val = apiStateRefs.rawData.value;
        if (!val) {
          return null;
        }
        return cb(val);
      });
    },
    computedList(prop, cb) {
      return computed(() => {
        let val = apiStateRefs.rawData.value;
        if (!val?.[prop]) {
          return [];
        }
        return val[prop].map((obj) => cb(obj));
      });
    },
    computedProp(prop, cb) {
      return computed(() => {
        let val = apiStateRefs.rawData.value;
        if (!val?.[prop]) {
          return null;
        }
        return cb(val[prop]);
      });
    },
  };
}

import { computed, reactive, toRefs, watch } from "vue";

import { useThrottleToggle } from "@/utils/throttle.js";

export function makeState() {
  const apiState = reactive({
    rawData: null,
    isLoading: false,
    error: null,
    didLoad: false,
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
        apiState.didLoad = true;
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
  };
}

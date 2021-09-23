import { reactive } from "@vue/composition-api";

import { useThrottleToggle } from "@/utils/throttle.js";

export function makeState() {
  const apiState = reactive({
    rawData: null,
    isLoading: false,
    error: null,
    didLoad: false,
    isLoadingThrottled: null,
  });

  apiState.isLoadingThrottled = useThrottleToggle(apiState, {
    prop: "isLoading",
  });

  return {
    apiState,

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

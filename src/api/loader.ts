import { computed, reactive, toRefs, watch } from "vue";
import type { ComputedRef, Ref, WatchSource } from "vue";

import { useThrottleToggle } from "@/utils/wait.ts";
import type { Result } from "@/utils/try-to.ts";

// CoreState holds the reactive fields for a single API request lifecycle.
interface LoadingState {
  rawData: unknown;
  isLoading: boolean;
  error: unknown;
}

export function makeState() {
  const apiState = reactive<LoadingState>({
    rawData: null,
    isLoading: false,
    error: null,
  });

  const apiStateRefs = toRefs(apiState);
  // isLoadingThrottled stays true for a short window after loading ends,
  // preventing loading-spinner flicker.
  const isLoadingThrottled = useThrottleToggle(apiStateRefs.isLoading);
  const apiStateRefsWithThrottle = { ...apiStateRefs, isLoadingThrottled };

  return {
    // @deprecated: use apiStateRefs instead; direct access to the reactive
    // object bypasses TypeScript and makes future migration harder.
    apiState,
    apiStateRefs: apiStateRefsWithThrottle,

    async exec<T>(callback: () => Promise<Result<T>>): Promise<void> {
      if (apiState.isLoading) {
        return;
      }
      apiState.isLoading = true;
      let data: unknown;
      [data, apiState.error] = (await callback()) as Result<unknown>;
      apiState.isLoading = false;
      if (!apiState.error) {
        apiState.rawData = data;
      }
    },
  };
}

export function watchAPI<T>(
  watchCb: WatchSource,
  fetcher: (val: unknown) => Promise<Result<T>>
) {
  const { exec, apiStateRefs } = makeState();
  const doFetch = (newVal: unknown) => exec(() => fetcher(newVal));

  watch(watchCb, doFetch, { immediate: true });

  // Typed accessor for rawData.
  const rawData = (): T | null => (apiStateRefs.rawData as Ref<T | null>).value;

  return {
    apiState: apiStateRefs,
    async fetch(): Promise<void> {
      return doFetch(typeof watchCb === "function" ? watchCb() : watchCb.value);
    },
    computer<R>(cb: (val: T | null) => R): ComputedRef<R> {
      return computed(() => cb(rawData()));
    },
    computedObj<R>(cb: (val: T) => R): ComputedRef<R | null> {
      return computed(() => {
        const val = rawData();
        if (!val) {
          return null;
        }
        return cb(val);
      });
    },
    computedList<K extends keyof T, R>(
      prop: K,
      cb: (obj: T[K] extends (infer I)[] ? I : never) => R
    ): ComputedRef<R[]> {
      return computed(() => {
        const val = rawData();
        const list = val?.[prop];
        if (!Array.isArray(list)) {
          return [];
        }
        return list.map(cb);
      });
    },
    computedProp<K extends keyof T, R>(
      prop: K,
      cb: (val: T[K]) => R
    ): ComputedRef<R | null> {
      return computed(() => {
        const val = rawData();
        if (!val?.[prop]) {
          return null;
        }
        return cb(val[prop]);
      });
    },
  };
}

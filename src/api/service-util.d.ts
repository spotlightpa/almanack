import type { ComputedRef, Ref } from "vue";

declare module "@/api/service-util.js" {
  interface ApiStateRefs {
    rawData: Ref<unknown>;
    isLoading: Ref<boolean>;
    error: Ref<unknown>;
    isLoadingThrottled: Ref<boolean>;
  }

  export function makeState(): {
    apiState: ReturnType<(typeof import("vue"))["reactive"]>;
    apiStateRefs: ApiStateRefs;
    exec(callback: () => Promise<[unknown, unknown]>): Promise<void>;
  };

  interface WatchAPIResult<T> {
    apiState: ApiStateRefs;
    fetch(): Promise<void>;
    computedList<U>(prop: string, cb: (item: T) => U): ComputedRef<U[]>;
    computedProp<U>(
      prop: string,
      cb: (val: unknown) => U
    ): ComputedRef<U | null>;
  }

  export function watchAPI<TKey, TItem>(
    watchCb: () => TKey,
    fetcher: (key: TKey) => Promise<[unknown, unknown]>
  ): WatchAPIResult<TItem>;
}

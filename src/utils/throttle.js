import { computed, ref, watch } from "@vue/composition-api";

export function useThrottleToggle(watchedRef, timeout = 1000) {
  const recentlyChanged = ref(false);
  watch(
    watchedRef,
    (val) => {
      if (val) {
        recentlyChanged.value = true;
        window.setTimeout(() => {
          recentlyChanged.value = false;
        }, timeout);
      }
    },
    { immediate: true }
  );
  return computed(() => watchedRef.value || recentlyChanged.value);
}

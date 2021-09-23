import { computed, ref, toRef, watch } from "@vue/composition-api";

export function useThrottleToggle(
  watchedRef,
  { prop = "", timeout = 1000 } = {}
) {
  if (prop) {
    watchedRef = toRef(watchedRef, prop);
  }
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

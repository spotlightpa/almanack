import { computed, ref, toRef, watch } from "vue";

export function useThrottleToggle(
  watchedRef,
  { prop = "", timeout = 1000 } = {}
) {
  const watched = prop ? toRef(watchedRef, prop) : watchedRef;

  const recentlyChanged = ref(false);
  watch(
    watched,
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

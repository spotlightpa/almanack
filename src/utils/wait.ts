import { computed, ref, watch, type Ref } from "vue";

type Timer = ReturnType<typeof window.setTimeout>;

// seconds converts a duration in seconds to milliseconds.
export const seconds = (s: number): number => 1_000 * s;

// wait returns a Promise that resolves after `ms` milliseconds.
export function wait(ms: number): Promise<void> {
  return new Promise((resolve) => {
    window.setTimeout(resolve, ms);
  });
}

// debounce wraps `fn` so it is only called after `ms` milliseconds of
// inactivity. Each new call resets the timer.
export function debounce<A extends unknown[]>(
  ms: number,
  fn: (...args: A) => void
) {
  let timer: Timer | undefined;
  return (...args: A) => {
    window.clearTimeout(timer);
    timer = window.setTimeout(() => fn(...args), ms);
  };
}

// useThrottleToggle keeps a ref true
// for at least `timeout` ms after the watched ref last became truthy.
// Useful for preventing loading-spinner flicker.
export function useThrottleToggle<T>(
  watchedRef: Ref<T>,
  { timeout = 1000 } = {}
) {
  const recentlyChanged = ref(false);
  let timer: Timer | undefined;
  watch(
    watchedRef,
    (val) => {
      if (val) {
        window.clearTimeout(timer);
        recentlyChanged.value = true;
        timer = window.setTimeout(() => {
          recentlyChanged.value = false;
        }, timeout);
      }
    },
    { immediate: true }
  );
  return computed(() => watchedRef.value || recentlyChanged.value);
}

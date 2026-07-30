import { computed, ref, watch, type Ref } from "vue";

// seconds converts a duration in seconds to milliseconds.
export const seconds = (s: number): number => 1000 * s;

// wait returns a Promise that resolves after `ms` milliseconds.
export function wait(ms: number): Promise<void> {
  return new Promise((resolve) => {
    window.setTimeout(resolve, ms);
  });
}

// debounce wraps `fn` so it is only called after `ms` milliseconds of
// inactivity. Each new call resets the timer.
export function debounce<T extends unknown[]>(
  fn: (...args: T) => void,
  ms: number
): (...args: T) => void {
  let timer: ReturnType<typeof window.setTimeout> | null = null;
  return function (this: unknown, ...args: T) {
    window.clearTimeout(timer ?? undefined);
    timer = window.setTimeout(() => fn.apply(this, args), ms);
  };
}

// useThrottleToggle keeps a ref true for at least `timeout` ms after the
// watched ref last became true. Useful for preventing loading-spinner flicker.
export function useThrottleToggle(
  watchedRef: Ref<boolean>,
  { timeout = 1000 } = {}
) {
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

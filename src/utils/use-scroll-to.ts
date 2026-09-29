import { ref, nextTick, type Ref } from "vue";

export default function useScrollTo(
  querystring = "[data-scroll-to]",
  position = -1
): [Ref<HTMLElement | undefined>, () => Promise<void>] {
  const container = ref<HTMLElement>();

  async function trigger(): Promise<void> {
    await nextTick();
    let el = container.value!;
    let headings = el.querySelectorAll(querystring);
    let newPick = Array.from(headings).at(position);
    newPick?.scrollIntoView({
      behavior: "smooth",
      block: "start",
    });
  }
  return [container, trigger];
}

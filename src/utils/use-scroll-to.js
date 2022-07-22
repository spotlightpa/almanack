import { ref, nextTick } from "vue";

export default function useScrollTo(
  querystring = "[data-scroll-to]",
  position = -1
) {
  const container = ref();

  async function trigger() {
    await nextTick();
    let el = container.value;
    let headings = el.querySelectorAll(querystring);
    let newPick = Array.from(headings).at(position);
    newPick.scrollIntoView({
      behavior: "smooth",
      block: "start",
    });
  }
  return [container, trigger];
}

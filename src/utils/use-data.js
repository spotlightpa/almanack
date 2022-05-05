import { ref, computed } from "@vue/composition-api";

export default function useData(getData, props) {
  const data = {};
  for (let [prop, [name, wrap = (v) => v, unwrap = (v) => v]] of Object.entries(
    props
  )) {
    const inner = ref(null);
    data[prop] = computed({
      get: () => inner.value ?? wrap(getData()[name]),
      set: (val) => {
        inner.value = val;
        getData()[name] = unwrap(val);
      },
    });
  }
  return data;
}

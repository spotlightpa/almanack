import { ref, computed } from "vue";

export default function useData(getData, props) {
  const data = {};
  let innerVals = [];
  for (let [prop, [name, wrap = (v) => v, unwrap = (v) => v]] of Object.entries(
    props
  )) {
    const inner = ref(null);
    innerVals.push(inner);
    data[prop] = computed({
      get: () => inner.value ?? wrap(getData()[name]),
      set: (val) => {
        inner.value = val;
        getData()[name] = unwrap(val);
      },
    });
  }
  data.resetData = () => {
    for (let val of innerVals) {
      val.value = null;
    }
  };
  return data;
}

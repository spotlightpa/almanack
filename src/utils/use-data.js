import { watch, reactive, toRefs } from "@vue/composition-api";

export default function useData(emit, src, props) {
  let data = {};
  for (let [prop, [name, wrap = (v) => v]] of Object.entries(props)) {
    data[prop] = wrap(src[name]);
  }
  data = toRefs(reactive(data));
  for (let [prop, [name, , unwrap = (v) => v]] of Object.entries(props)) {
    watch(data[prop], (val) => {
      emit("data", [name, unwrap(val)]);
    });
  }
  return data;
}

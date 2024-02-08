import { ref } from "vue";

export default function useProps(src, props) {
  const data = {};
  for (let [prop, [name, wrap = (v) => v]] of Object.entries(props)) {
    data[prop] = ref(wrap(src[name]));
  }
  function saveData() {
    let dst = {};
    for (let [prop, [name, , unwrap = (v) => v]] of Object.entries(props)) {
      dst[name] = unwrap(data[prop].value);
    }
    return dst;
  }
  return [data, saveData];
}

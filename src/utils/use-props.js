import { ref } from "vue";

// useProps takes an object and turns select props into refs.
// The mapping is an object whose keys are turned into names on returned data object.
// The values of the mapping are the names of the props on source object
// and then optional transformations to do on deserialize and serialize.
export default function useProps(src, mapping) {
  const data = {};
  for (let [prop, [name, deserialize = (v) => v]] of Object.entries(mapping)) {
    data[prop] = ref(deserialize(src[name]));
  }
  function saveData() {
    let dst = {};
    for (let [prop, [name, , serialize = (v) => v]] of Object.entries(
      mapping
    )) {
      dst[name] = serialize(data[prop].value);
    }
    return dst;
  }
  return [data, saveData];
}

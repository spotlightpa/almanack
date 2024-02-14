<script setup>
import useProps from "@/utils/use-props.js";

const props = defineProps({
  params: Object,
  propName: String,
  fileProps: Object,
});
let n = 0;

function deserialize(v) {
  let a = v || [];
  return a.map((o) => ({ ...o, id: n++ }));
}

function serialize(v) {
  return v.map((o) => ({ ...o, id: undefined }));
}

let [{ imageSet }, saveData] = useProps(props.params.data, {
  imageSet: [props.propName, deserialize, serialize],
});

function pushImage() {
  imageSet.value.push({
    id: n++,
    link: "https://www.spotlightpa.org/donate/",
    description: "Lorem ipsum",
    sources: [],
  });
}

defineExpose({
  saveData,
});
</script>

<template>
  <ul>
    <li v-for="image of imageSet" :key="image.id" class="zebra-row">
      <BulmaFieldInput
        v-model="image.link"
        label="Link URL"
        type="url"
        placeholder="https://www.spotlightpa.org/donate/"
      />
      <BulmaTextarea
        v-model="image.description"
        label="Image description (alt text)"
        help="For blind readers and search engines"
      />
      <BulmaField
        label="Images"
        help="If multiple images are provided, each page load will select one randomly"
      >
        <SiteParamsFiles
          :files="image.sources"
          :file-props="fileProps"
          @add="image.sources.push($event)"
          @remove="image.sources.splice($event, 1)"
        />
      </BulmaField>
    </li>
  </ul>
  <div class="buttons">
    <button
      type="button"
      class="button is-primary has-text-weight-semibold"
      @click="pushImage"
    >
      Add Image
    </button>
  </div>
</template>

<style scoped>
.zebra-row {
  background-color: #fff;
}

.zebra-row:nth-child(even) {
  background-color: #fafafa;
}

.zebra-row + .zebra-row {
  border-top: 1px solid #dbdbdb;
}
</style>
